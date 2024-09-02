package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/glueckkanja/terraform-provider-aztools/internal/provider/common"
	"github.com/glueckkanja/terraform-provider-aztools/internal/provider/models"
	"github.com/glueckkanja/terraform-provider-aztools/internal/provider/validate"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceName() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNameCreate,
		ReadContext:   resourceNameRead,
		// UpdateContext: All fields are ForceNew or Computed w/out Optional, Update is superfluous
		DeleteContext: resourceNameDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"convention": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if !(v == "default" || v == "passthrough") {
						errs = append(errs, fmt.Errorf("allowed convention values are 'default' or 'passthrough'"))
					}
					return
				},
			},
			"prefixes": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
				Optional: true,
				ForceNew: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// ValidateFunc: validation.StringDoesNotContainAny(), // FIXME: Add validation
				Default: nil,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				//ValidateFunc: validation.NoZeroValues, // FIXME: Add validation
				Default: nil,
			},
			"hash_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 9999),
				Default:      0,
			},
			"suffixes": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
				Optional: true,
				ForceNew: true,
			},
			"separator": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				//ValidateFunc: validation.StringIsNotWhiteSpace,  // FIXME: Add validation
				Default: nil,
			},
			"name_precedence": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
				Optional: true,
				ForceNew: true,
			},
			"result": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
		},
		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIfChange("hash_length", func(_ context.Context, old, new, _ interface{}) bool {
				// "size" can only increase in-place, so we must create a new resource
				// if it is decreased.
				return new.(int) != old.(int)
			}),
		),
	}
}

/*
func valHash(old, new, meta interface{}) bool {
	// "size" can only increase in-place, so we must create a new resource
	// if it is decreased.
	return new.(int) < old.(int)
}
*/

func resourceNameCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	diags := resourceNameRead(ctx, d, meta)

	// If resourceNameRead returned diadnostics error, return
	if diags.HasError() {
		return diags
	}

	d.SetId(uuid.New().String())

	return diags
}

func resourceNameDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceNameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	providerConfig := meta.(*models.ProviderConfiguration)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	//----------------------------------------------------------------------------------------------------
	// Validate resource resourceData 'resource_type' using 'ResourceDefinition' Schema and return Schema

	// Define schema for result object
	type resultSchema struct {
		NamingSchema          models.NamingSchema // Naming Schema for resource_type
		ResourceConfiguration struct {            // Computed Result
			Name           string
			ResourceType   string
			Convention     string
			Environment    string
			Location       string
			HashLength     int
			Prefixes       []string
			Suffixes       []string
			Separator      string
			NamePrecedence []string
			ForceResfresh  bool
		}
		Result string
	}

	result := resultSchema{}

	// resource_type
	// TODO: Add error handling
	if i, ok := d.GetOk("resource_type"); ok {
		result.ResourceConfiguration.ResourceType = i.(string)
	}

	//----------------------------------------------------------------------------------------------------
	// Find naming schema for specific resource_type by key and store it in result object

	if v, ok := providerConfig.NamingSchemaMap[result.ResourceConfiguration.ResourceType]; ok {
		// Found
		result.NamingSchema = v
	} else {
		// Not found
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find resource_type '" + result.ResourceConfiguration.ResourceType + "'",
			Detail:   "Can not find '" + result.ResourceConfiguration.ResourceType + "' definition in json naming schema file.",
		})
		return diags
	}

	// Convention - if 'convention' specified
	result.ResourceConfiguration.Convention = providerConfig.Convention
	// Environment - overrrided by resource configuration
	if i, ok := d.GetOk("convention"); ok {
		if i != nil {
			result.ResourceConfiguration.Convention = i.(string)
		}
	}

	// Name
	// TODO: Add error handling
	if i, ok := d.GetOk("name"); ok {
		result.ResourceConfiguration.Name = i.(string)
	}

	// If naming convention == 'passthrough' then skip calculation
	if result.ResourceConfiguration.Convention != "passthrough" {

		//----------------------------------------------------------------------------------------------------
		// Get configuration from resurce or if not configured then get it from provider configuration/default values

		// Environment - if 'UseEnvironment' enabled
		if result.NamingSchema.Configuration.UseEnvironment {
			result.ResourceConfiguration.Environment = providerConfig.Environment
		} else {
			result.ResourceConfiguration.Environment = ""
		}
		// Environment - overrrided by resource configuration
		if i, ok := d.GetOk("environment"); ok {
			if i != nil {
				result.ResourceConfiguration.Environment = i.(string)
			}
		}

		//----------------------------------------------------------------------------------------------------
		// location

		// priority selection
		// 1 - resource

		if i, ok := d.GetOk("location"); ok {
			if i != nil {
				if v, ok := providerConfig.LocationsMap[i.(string)]; ok {
					// Found
					result.ResourceConfiguration.Location = v
				} else {
					// Not found
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Unable to find location '" + i.(string) + "'",
						Detail:   "Can not find '" + i.(string) + "' value in json locations schema file.",
					})
					return diags
				}
			}
		}

		//----------------------------------------------------------------------------------------------------
		// separator

		// priority selection
		// 1 - resource
		// 2 - provider configuration or default

		if result.NamingSchema.Configuration.UseSeparator {
			result.ResourceConfiguration.Separator = providerConfig.Separator
		}
		if i, ok := d.GetOk("separator"); ok {
			if i != nil {
				result.ResourceConfiguration.Separator = i.(string)
			}
		}

		//----------------------------------------------------------------------------------------------------
		// name precedence

		// TODO: implement provider configuration

		// priority selection
		// 1 - resource
		// 2 - provider configuration - NOT YET IMPLEMENTED
		// 2 - json schema
		// 3 - default

		result.ResourceConfiguration.NamePrecedence = []string{"abbreviation", "prefixes", "name", "location", "environment", "hash", "suffixes"}
		// NamePrecedence - If schema contain NamePrecedence values
		if len(result.NamingSchema.Configuration.NamePrecedence) > 0 {
			result.ResourceConfiguration.NamePrecedence = result.NamingSchema.Configuration.NamePrecedence
		}
		// NamePrecedence - overrrided by resource configuration
		if i, ok := d.GetOk("name_precedence"); ok {
			result.ResourceConfiguration.NamePrecedence = common.ConvertInterfaceToString(i.([]interface{}))
		}

		//----------------------------------------------------------------------------------------------------
		// preffixes and suffixes

		// TODO: Add error handling

		// priority selection
		// 1 - resource

		// Prefixes
		if i, ok := d.GetOk("prefixes"); ok {
			result.ResourceConfiguration.Prefixes = common.ConvertInterfaceToString(i.([]interface{}))
		}
		// Suffixes
		if i, ok := d.GetOk("suffixes"); ok {
			result.ResourceConfiguration.Suffixes = common.ConvertInterfaceToString(i.([]interface{}))
		}

		//----------------------------------------------------------------------------------------------------
		// Compute random hash

		// priority selection
		// 1 - resource
		// 2 - provider configuration
		// 3 - json schema

		// if 'hash_length' parameter defined on resource
		if i, ok := d.GetOk("hash_length"); ok {
			result.ResourceConfiguration.HashLength = i.(int)
		} else {
			// if 'hash_length' parameter defined in provide configuration
			if providerConfig.HashLength != 0 {
				result.ResourceConfiguration.HashLength = providerConfig.HashLength
			} else {
				// if 'hashLength' parameter defined in json schema
				if result.NamingSchema.Configuration.HashLength != 0 {
					result.ResourceConfiguration.HashLength = result.NamingSchema.Configuration.HashLength
				} else {
					// If none above
					result.ResourceConfiguration.HashLength = 0
				}
			}
		}

		randomSeed := int64(result.ResourceConfiguration.HashLength)
		randomHash := common.RandomSeq(result.ResourceConfiguration.HashLength, &randomSeed)

		//----------------------------------------------------------------------------------------------------
		// Compute resourceNameResult

		calculatedContent := []string{}

		for i := 0; i < len(result.ResourceConfiguration.NamePrecedence); i++ {
			switch c := result.ResourceConfiguration.NamePrecedence[i]; c {
			case "abbreviation":
				if len(result.NamingSchema.Abbreviation) > 0 {
					calculatedContent = append(calculatedContent, result.NamingSchema.Abbreviation)
				}
			case "prefixes":
				if len(result.ResourceConfiguration.Prefixes) > 0 {
					if len(result.ResourceConfiguration.Prefixes[0]) > 0 {
						calculatedContent = append(calculatedContent, result.ResourceConfiguration.Prefixes...)
					}
					result.ResourceConfiguration.Prefixes = result.ResourceConfiguration.Prefixes[1:]
					if len(result.ResourceConfiguration.Prefixes) > 0 {
						i--
					}
				}
			case "name":
				if len(result.ResourceConfiguration.Name) > 0 {
					calculatedContent = append(calculatedContent, result.ResourceConfiguration.Name)
				}
			case "environment":
				if len(result.ResourceConfiguration.Environment) > 0 {
					calculatedContent = append(calculatedContent, result.ResourceConfiguration.Environment)
				}
			case "location":
				if len(result.ResourceConfiguration.Location) > 0 {
					calculatedContent = append(calculatedContent, result.ResourceConfiguration.Location)
				}
			case "hash":
				if result.ResourceConfiguration.HashLength > 0 {
					calculatedContent = append(calculatedContent, randomHash)
				}
			case "suffixes":
				if len(result.ResourceConfiguration.Suffixes) > 0 {
					if len(result.ResourceConfiguration.Suffixes[0]) > 0 {
						calculatedContent = append(calculatedContent, result.ResourceConfiguration.Suffixes[0])
					}
					result.ResourceConfiguration.Suffixes = result.ResourceConfiguration.Suffixes[1:]
					if len(result.ResourceConfiguration.Suffixes) > 0 {
						i--
					}
				}
			}
		}

		result.Result = strings.Join(calculatedContent, result.ResourceConfiguration.Separator)
	} else {
		// Send name as 'passthrough'
		result.Result = result.ResourceConfiguration.Name
	}

	// Use lowercase if set to true before validation
	if result.NamingSchema.Configuration.UseLowerCase {
		result.Result = strings.ToLower(result.Result)
	}

	//----------------------------------------------------------------------------------------------------
	// Validate resourceNameResult

	// Validate resourceName against MaxLength
	if ok, diags := validate.Length(result.Result, result.NamingSchema.ResourceType, result.NamingSchema.MaxLength); !ok {
		return diags
	}

	// Validate resourceName against ValidationRegex
	if ok, diags := validate.RegEx(result.Result, result.NamingSchema.ResourceType, result.NamingSchema.ValidationRegex); !ok {
		return diags
	}

	// Validate resourceName contains double hyphens
	if result.NamingSchema.Configuration.DenyDoubleHyphens {
		if ok, diags := validate.DoubleHyphens(result.Result, result.ResourceConfiguration.ResourceType); !ok {
			return diags
		}
	}

	//// Validate resourceName against regex
	if value, ok := d.Get("resource_type").(string); ok {
		result.ResourceConfiguration.ResourceType = value
	}

	//----------------------------------------------------------------------------------------------------
	// Set result

	//d.Set("result", result.Result)
	err := d.Set("result", result.Result)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error in AF naming result for resource '" + result.ResourceConfiguration.ResourceType + "'",
			Detail:   "Unknown error for resourceType '" + result.ResourceConfiguration.ResourceType + "' in result '" + result.ResourceConfiguration.Name + "'. Result '" + result.Result + "' not alowed.",
		})
		return diags
	}

	return diags
}
