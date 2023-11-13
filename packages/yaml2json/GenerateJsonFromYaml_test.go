package yaml2json

import (
	"testing"
)

// TestMyFunc tests the function myfunc with multiple different inputs.
func TestGenerateJsonFromYaml(t *testing.T) {
	// Define test cases
	tests := []struct {
		testName    string
		yamlBytes   []byte
		jsonBytes   []byte
		expectedErr bool
	}{
		// Empty string
		{
			testName:    "Empty string",
			yamlBytes:   []byte(``),
			jsonBytes:   []byte(``),
			expectedErr: true,
		},
		// Basic working example
		{
			testName: "Basic working example",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: object
      _title_and_description_in_markdown: "{{ self.alias }}"
      _format: grid

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "{{ self.alias }}",
						"title": " ",
						"type": "object",
						"format": "grid",
						"description": "",
						"properties": {}
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Catalog_title_and_description_in_markdown - empty
		{
			testName: "Catalog_title_and_description_in_markdown - empty",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: ""

Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: object
      _title_and_description_in_markdown: "{{ self.alias }}"
      _format: grid

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "",
	"description": "",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "{{ self.alias }}",
						"title": " ",
						"type": "object",
						"format": "grid",
						"description": "",
						"properties": {}
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Catalog_title_and_description_in_markdown - missing description
		{
			testName: "Catalog_title_and_description_in_markdown - missing description",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: object
      _title_and_description_in_markdown: "{{ self.alias }}"
      _format: grid

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "{{ self.alias }}",
						"title": " ",
						"type": "object",
						"format": "grid",
						"description": "",
						"properties": {}
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Catalog_products {} error
		{
			testName: "Catalog_products {}",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {}
}
`),
			expectedErr: true,
		},
		// ProductName with invalid-chars, error
		{
			testName: "ProductName with invalid-chars, error",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "invalid!chars#":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: object
      _title_and_description_in_markdown: "{{ self.alias }}"
      _format: grid

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "{{ self.alias }}",
						"title": " ",
						"type": "object",
						"format": "grid",
						"description": "",
						"properties": {}
					}
				}
			}
		}
	}
}
`),
			expectedErr: true,
		},
		// Product_nr_of_items invalid, error
		{
			testName: "Product_nr_of_items invalid, error",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: junk 

    Product_item:
      _type: object
      _title_and_description_in_markdown: "{{ self.alias }}"
      _format: grid

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "{{ self.alias }}",
						"title": " ",
						"type": "object",
						"format": "grid",
						"description": "",
						"properties": {}
					}
				}
			}
		}
	}
}
`),
			expectedErr: true,
		},
		// Product_item object
		{
			testName: "Product_item object",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: object
      _title_and_description_in_markdown: "{{ self.alias }}"
      _format: grid
      "mypropX":
        _type: boolean

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "{{ self.alias }}",
						"title": " ",
						"type": "object",
						"format": "grid",
						"description": "",
						"properties": {
                            "mypropX": {
                                "description": "",
                                "format": "checkbox",
                                "headerTemplate": "",
                                "title": " ",
                                "type": "boolean"
                            }							
						}
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// object _zzzz to zzzz
		{
			testName: "object _zzzz to zzzz",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: object
      _title_and_description_in_markdown: "{{ self.alias }}"
      _format: grid
      _extra_field: 
        xa: 1
        _xb: [2,2,2]
        xc: true
`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "{{ self.alias }}",
						"title": " ",
						"type": "object",
						"format": "grid",
						"extra_field": { 
					    	"xa": 1,
						 	"_xb": [2,2,2],
							"xc": true
						},
						"description": "",
						"properties": {}
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},

		// object _format missing is set to null
		{
			testName: "object _format missing is set to null",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: object
      _title_and_description_in_markdown: "{{ self.alias }}"
    # _format: grid
`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "{{ self.alias }}",
						"title": " ",
						"type": "object",
						"format": null,
						"description": "",
						"properties": {}
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},

		// Product_item string
		{
			testName: "Product_item string",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: string
      _title_and_description_in_markdown: |-
        Internal alias
        --------------------
        An internal alias to remember what this object_id corresponds to


      _default: "Exs: 'group TeamWhisky' or 'user Siegi' or 'spn cicdWhisky1'"

    # [opt] one of these
      _pattern: '^[0-9a-z- ]+$'
    # _format: [color|date|datetime-local|email|month|password|number|range|tel|text|textarea|time|url|week]
    # _format: <je:specialized string editor + its options>   >> https://github.com/json-editor/json-editor/blob/master/README.md#specialized-string-editors
    # _enum: ["opt1", "opt2", "opt3"]     # [opt] enum deactivates other validating options like "format", ...

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "Internal alias",
						"title": " ",
						"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
						"type": "string",
						"pattern": "^[0-9a-z- ]+$",
						"default": "Exs: 'group TeamWhisky' or 'user Siegi' or 'spn cicdWhisky1'"
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Product_item bool
		{
			testName: "Product_item bool",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: boolean
      _title_and_description_in_markdown: |-
        Internal alias
        --------------------
        An internal alias to remember what this object_id corresponds to

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "Internal alias",
						"title": " ",
						"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
						"format": "checkbox",
						"type": "boolean"
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Product_item bool default true
		{
			testName: "Product_item bool default true",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: boolean
      _title_and_description_in_markdown: |-
        Internal alias
        --------------------
        An internal alias to remember what this object_id corresponds to

      _default: true

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "Internal alias",
						"title": " ",
						"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
						"format": "checkbox",
						"type": "boolean",
						"default": true
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Product_item number
		{
			testName: "Product_item number",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: number
      _title_and_description_in_markdown: |-
        Internal alias
        --------------------
        An internal alias to remember what this object_id corresponds to


`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "Internal alias",
						"title": " ",
						"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
						"type": "number"
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Product_item integer
		{
			testName: "Product_item integer",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: integer
      _title_and_description_in_markdown: |-
        Internal alias
        --------------------
        An internal alias to remember what this object_id corresponds to


`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "Internal alias",
						"title": " ",
						"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
						"type": "integer"
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Product_item integer enum
		{
			testName: "Product_item integer enum",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: integer
      _title_and_description_in_markdown: |-
        Internal alias
        --------------------
        An internal alias to remember what this object_id corresponds to

      _enum: [3, 5]
`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "Internal alias",
						"title": " ",
						"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
						"type": "integer",
						"enum": [3,5]
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Product_item array
		{
			testName: "Product_item array",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: array
      _title_and_description_in_markdown: |-
        Internal alias
        --------------------
        An internal alias to remember what this object_id corresponds to

      _items:
        _type: boolean
        _title_and_description_in_markdown: |-
          Internal alias
          --------------------
          An internal alias to remember what this object_id corresponds to

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "Internal alias",
						"title": " ",
						"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
						"type": "array",
						"options": {
							"collapsed": false
						},
						"minItems": 0, "maxItems": 1000,
						"items": {
							"headerTemplate": "Internal alias",
							"title": " ",
							"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
							"format": "checkbox",
							"type": "boolean"
						}
					}
				}
			}
		}
	}
}
`),
			expectedErr: false,
		},
		// Product_item array items empty, error
		{
			testName: "Product_item array items empty, error",
			yamlBytes: []byte(`
Catalog_title_and_description_in_markdown: |-
    Select products
    --------------------
    See our wiki **bold**


Catalog_products:
  "additional_members_of_AADGroup_DataClients":
    Product_title_and_description_in_markdown: |-
        AAD-Group 'mycollection_DataClients' members
        --------------------
        The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files
        from storage accounts, secrets from keyvaults, etc

    Product_nr_of_items: 0-inf 

    Product_item:
      _type: array
      _title_and_description_in_markdown: |-
        Internal alias
        --------------------
        An internal alias to remember what this object_id corresponds to

      _items: {}

`),
			jsonBytes: []byte(`
{
	"headerTemplate": "Select products",
	"description": "\u003cp\u003eSee our wiki \u003cstrong\u003ebold\u003c/strong\u003e\u003c/p\u003e\n",
	"title": " ",
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
		"additional_members_of_AADGroup_DataClients": {
			"propertyOrder": 1001,
			"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
			"title": " ",
			"type": "object",
			"options": {
				"collapsed": true
			},
			"properties": {
				"info": {
					"propertyOrder": 1,
					"description": "\u003cp\u003eThe \u0026lsquo;mycollection_DataClients\u0026rsquo; AAD-Group members can access data from resources: blob/files\nfrom storage accounts, secrets from keyvaults, etc\u003c/p\u003e\n",
					"type": "info"
				},
				"elements": {
					"propertyOrder": 2,
					"title": " ",
					"type": "array",
					"options": {
						"collapsed": false,
						"disable_collapse": true
					},
					"maxItems": 1000,
					"minItems": 0,
					"format": "tabs",
					"items": {
						"headerTemplate": "Internal alias",
						"title": " ",
						"description": "\u003cp\u003eAn internal alias to remember what this object_id corresponds to\u003c/p\u003e\n",
						"type": "array",
						"options": {
							"collapsed": false
						},
						"minItems": 0, "maxItems": 1000,
						"items": {}
					}
				}
			}
		}
	}
}
`),
			expectedErr: true,
		},

		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Call the function with the test case input
			got, err := GenerateJsonFromYaml(tt.yamlBytes)

			// Check if error is not nil when it's expected to be not nil
			{
				returnedErr := (err != nil)
				if returnedErr != tt.expectedErr {
					t.Fatalf("GenerateJsonFromYaml() error = %v, expectedErr %v", err, tt.expectedErr)
					return
				}
				if returnedErr {
					return
				}
			}
			// Check if the function returns the expected output
			gotPp, err := jsonPrettyPrinter(string(got))
			if err != nil {
				t.Fatalf("unexpected error %s", err)
			}
			desiredPp, err := jsonPrettyPrinter(string(tt.jsonBytes))
			if err != nil {
				t.Fatalf("unexpected error %s", err)
			}
			if gotPp != desiredPp {
				t.Fatalf("GenerateJsonFromYaml() = %v, want %v", gotPp, desiredPp)
			}
		})
	}
}
