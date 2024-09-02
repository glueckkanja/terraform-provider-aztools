package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/glueckkanja/terraform-provider-aztools/internal/provider/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"
)

// AzTools -
func AzTools(version string) func() *schema.Provider {
	return func() *schema.Provider {

		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"convention": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_NAMING_CONVENTION", "default"),
					Description: "Default convention for all naming results. Possible values 'default', 'passthrough'. Default 'default'",
					ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
						v := val.(string)
						if !(v == "default" || v == "passthrough") {
							errs = append(errs, fmt.Errorf("provider configuration: allowed convention values are 'default' or 'passthrough'"))
						}
						return
					},
				},
				"environment": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_NAMING_ENVIRONMENT", "sandbox"),
					Description: "Environment parameter. Default 'sandbox'",
				},
				"separator": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_NAMING_SEPARATOR", "-"),
					Description: "Naming schema separator. Default '-'",
				},
				"hash_length": {
					Type:        schema.TypeInt,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_HASH_LENGTH", nil),
					Description: "Default hash length for resource schema. Overrrides all naminng schema configurations defined in json files.",
				},
				"lowercase": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_NAMING_LOWERCASE", false),
					Description: "Namig result formating. Default 'false'",
				},
				"schema_naming_path": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"schema_naming_url"},
					DefaultFunc:   schema.EnvDefaultFunc("AZF_NAMING_JSON_FILE_PATH", nil),
					Description:   "Path to the config file.",
				},
				"schema_naming_url": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"schema_naming_path"},
					DefaultFunc:   schema.EnvDefaultFunc("AZF_NAMING_JSON_URL", nil),
					Description:   "Path to the config file.",
				},
				"schema_locations_path": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"schema_locations_url"},
					DefaultFunc:   schema.EnvDefaultFunc("AZF_LOCATIONS_JSON_FILE_PATH", nil),
					Description:   "Path to the config file.",
				},
				"schema_locations_url": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"schema_locations_path"},
					DefaultFunc:   schema.EnvDefaultFunc("AZF_LOCATIONS_JSON_URL", nil),
					Description:   "Url to the config file.",
				},
			},

			DataSourcesMap: map[string]*schema.Resource{
				// "scaffolding_data_source": dataSourceScaffolding(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"aztools_resource_name": resourceName(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

		// Warning or errors can be collected in a slice type
		var diags diag.Diagnostics

		var providerConfig models.ProviderConfiguration

		// Setup a User-Agent for your API client (replace the provider name for yours):
		providerConfig.UserAgent = p.UserAgent("terraform-provider-aztools", version)

		// Overriding with static configuration
		if i, ok := d.GetOk("convention"); ok {
			providerConfig.Convention = i.(string)
		}

		// Overriding with static configuration
		if i, ok := d.GetOk("environment"); ok {
			providerConfig.Environment = i.(string)
		}
		// TODO: add option to disable_separator?
		if i, ok := d.GetOk("separator"); ok {
			providerConfig.Separator = i.(string)
		}
		if i, ok := d.GetOk("lowercase"); ok {
			providerConfig.Lowercase = i.(bool)
		}

		if i, ok := d.GetOk("hash_length"); ok {
			providerConfig.HashLength = i.(int)
		}

		// load & parse naming schemas

		parsedDataNamingSchema, err := parseNamingJsonSchema(d)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		// Build a config map with 'ResourceType' as key:
		providerConfig.NamingSchemaMap = map[string]models.NamingSchema{}
		for _, v := range parsedDataNamingSchema {
			providerConfig.NamingSchemaMap[v.ResourceType] = v
		}

		providerConfig.LocationsMap, err = parseLocationsJsonSchema(d)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return &providerConfig, diags
	}
}

func parseNamingJsonSchema(d *schema.ResourceData) ([]models.NamingSchema, error) {

	var byteValue []byte
	var err error

	if configPath, ok := d.GetOk("schema_naming_path"); ok && configPath.(string) != "" {

		// read our jsonFile from file
		byteValue, err = readJSONFromFilePath(configPath.(string))
		if err != nil {
			return nil, err
		}

	} else if configUrl, ok := d.GetOk("schema_naming_url"); ok && configUrl.(string) != "" {

		// read our jsonFile from url
		byteValue, err = readJSONFromUrl(configUrl.(string))
		if err != nil {
			return nil, err
		}

	} else {

		err = fmt.Errorf("provider configuration: failed to read json schema configuration file from '" + configUrl.(string) + "'")
		return nil, err
	}

	// we initialize our NamingSchema array
	var namingSchema []models.NamingSchema

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'LocationsMapSchema' which we defined above
	if err := json.Unmarshal(byteValue, &namingSchema); err != nil {
		return nil, err
	}

	return namingSchema, nil
}

func parseLocationsJsonSchema(d *schema.ResourceData) (models.LocationsMapSchema, error) {

	var byteValue []byte
	var err error

	if configPath, ok := d.GetOk("schema_locations_path"); ok && configPath.(string) != "" {

		// read our jsonFile from file
		byteValue, err = readJSONFromFilePath(configPath.(string))
		if err != nil {
			return nil, err
		}

	} else if configUrl, ok := d.GetOk("schema_locations_url"); ok && configUrl.(string) != "" {

		// read our jsonFile from url
		byteValue, err = readJSONFromUrl(configUrl.(string))
		if err != nil {
			return nil, err
		}

	} else {

		err = fmt.Errorf("provider configuration: failed to read json schema configuration file from '" + configUrl.(string) + "'")
		return nil, err
	}

	// we initialize our LocationsSchema array
	var locationsMap models.LocationsMapSchema

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'LocationsMapSchema' which we defined above
	if err := json.Unmarshal(byteValue, &locationsMap); err != nil {
		return nil, err
	}

	return locationsMap, nil
}

func readJSONFromFilePath(path string) ([]byte, error) {

	// expand local path
	filePath, err := homedir.Expand(path)
	if err != nil {
		return nil, err
	}

	// Open our jsonFile
	jsonFile, err := os.Open(filePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	return byteValue, nil
}

func readJSONFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	var byteValue []byte

	if _, err := buf.ReadFrom(resp.Body); err == nil {
		byteValue = buf.Bytes()
	} else {
		return nil, err
	}

	return byteValue, nil
}
