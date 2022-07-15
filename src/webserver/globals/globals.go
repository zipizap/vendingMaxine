package globals

var (
	// TODO: this should not stay hardcoded like this... should come from env-var...
	Secret = []byte("my32bytesSuper$3cr37keyforCookies")
	//               123456789012343567890123456789012

	DebugDisableLogin = true
)

const Userkey = "user"
const PRODUCT_SCHEMA_JSON_FILEPATH = "config/products.schema.json"
