
package pkg

type Specification struct {
    Debug          bool `default:false`
    Port           int `required:"true", default: 9092`
    NamespaceLabel string `required:"true"`
	NamespaceClaim string `required:"true"`
	AdminRole      string
	VerifyToken    bool `default:false`
	JwksUrl        string `required:"true"`
	PrometheusUrl  string `required:"true"`
}