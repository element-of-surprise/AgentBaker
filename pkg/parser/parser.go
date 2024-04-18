package parser

import (
	"bytes"
	_ "embed"
	"log"
	"text/template"

	nbcontractv1 "github.com/Azure/agentbaker/pkg/proto/nbcontract/v1"
)

var (
	//go:embed cse_cmd.sh.gtpl
	bootstrapTrigger         string
	bootstrapTriggerTemplate = template.Must(template.New("triggerBootstrapScript").Funcs(getFuncMap()).Parse(bootstrapTrigger)) //nolint:gochecknoglobals

)

func executeBootstrapTemplate(inputContract *nbcontractv1.Configuration) (string, error) {
	var buffer bytes.Buffer
	if err := bootstrapTriggerTemplate.Execute(&buffer, inputContract); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// this function will eventually take a pointer to the bootstrap contract struct.
// it will then template out the variables into the final bootstrap trigger script.
func Parse(nbc *nbcontractv1.Configuration) {
	triggerBootstrapScript, err := executeBootstrapTemplate(nbc)
	if err != nil {
		log.Printf("Failed to execute the template: %v", err)
	}

	log.Println("output env vars:")
	log.Println(triggerBootstrapScript)
}
