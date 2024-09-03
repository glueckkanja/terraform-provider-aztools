package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
					Description: "",
					ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
						v := val.(string)
						if !(v == "default" || v == "passthrough") {
							errs = append(errs, fmt.Errorf("Provider configuration: Allowed convention values are 'default' or 'passthrough'"))
						}
						return
					},
				},
				"environment": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_NAMING_ENVIRONMENT", "sandbox"),
					Description: "",
				},
				"separator": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_NAMING_SEPARATOR", "-"),
					Description: "",
				},
				"lowercase": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_NAMING_LOWERCASE", false),
					Description: "",
				},
				"schema_naming_path": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_NAMING_JSON_FILE_PATH", "./schema.naming.json"),
					Description: "Path to the config file, defaults to ./schema.naming.json",
				},
				"schema_locations_path": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZF_LOCATIONS_JSON_FILE_PATH", "./schema.locations.json"),
					Description: "Path to the config file, defaults to ./schema.locations.json",
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

		// load & parse naming schema
		if configPath, ok := d.GetOk("schema_naming_path"); ok && configPath.(string) != "" {
			path, err := homedir.Expand(configPath.(string))
			if err != nil {
				return nil, diag.FromErr(err)
			}

			parsedData, err := parseJSONNamingSchema(path)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			// Build a config map with 'ResourceType' as key:
			providerConfig.NamingSchemaMap = map[string]models.NamingSchema{}
			for _, v := range parsedData {
				providerConfig.NamingSchemaMap[v.ResourceType] = v
			}
		}

		// load & parse locations schema
		if configPath, ok := d.GetOk("schema_locations_path"); ok && configPath.(string) != "" {
			path, err := homedir.Expand(configPath.(string))
			if err != nil {
				return nil, diag.FromErr(err)
			}

			providerConfig.LocationsMap, err = parseJSONLocationsMapSchema(path)
			if err != nil {
				return nil, diag.FromErr(err)
			}
		}

		return &providerConfig, diags
	}
}

func parseJSONNamingSchema(fp string) ([]models.NamingSchema, error) {

	// Open our jsonFile
	jsonFile, err := os.Open(fp)
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

	// we initialize our NamingSchema array
	var namingSchema []models.NamingSchema

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'namingSchema' which we defined above
	err = json.Unmarshal(byteValue, &namingSchema)
	if err != nil {
		return nil, err
	}

	return namingSchema, nil
}

func parseJSONLocationsMapSchema(fp string) (models.LocationsMapSchema, error) {

	// we initialize our NamingSchema array
	var locationsMap models.LocationsMapSchema

	// Open our jsonFile
	jsonFile, err := os.Open(fp)
	// if we os.Open returns an error then handle it
	if err != nil {
		return locationsMap, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return locationsMap, err
	}

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'LocationsMapSchema' which we defined above
	err = json.Unmarshal(byteValue, &locationsMap)
	if err != nil {
		return locationsMap, err
	}

	return locationsMap, nil
}
