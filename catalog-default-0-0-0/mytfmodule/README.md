# Root-module provider config
In your root-module, include something similar to the following provider config

```  
terraform {
  required_providers {
    azurerm = {
      source = "hashicorp/azurerm"
      version = "3.65.0"
    }
  }
}

provider "azurerm" {
  # Configuration options
}
```