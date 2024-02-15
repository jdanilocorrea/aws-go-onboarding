package add_policy

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Função para anexar uma política ao usuário pelo nome da política
func AttachPolicyByName(sess *session.Session, userName string, policyName string) {
	svc := iam.New(sess)

	// Obtém o ARN da política pelo nome
	arn, err := GetPolicyARNByName(svc, policyName)
	if err != nil {
		log.Fatalf("Erro ao obter ARN da política: %v", err)
	}

	// Anexa a política ao usuário
	_, err = svc.AttachUserPolicy(&iam.AttachUserPolicyInput{
		UserName:  aws.String(userName),
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		log.Fatalf("Erro ao anexar política ao usuário: %v", err)
	}

	fmt.Println("Política adicionada com sucesso ao usuário:", userName)
}

// Função para obter o ARN da política pelo nome
func GetPolicyARNByName(svc *iam.IAM, policyName string) (string, error) {
	input := &iam.ListPoliciesInput{}

	var policyArn string

	err := svc.ListPoliciesPages(input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		for _, policy := range page.Policies {
			if *policy.PolicyName == policyName {
				policyArn = *policy.Arn
				return false // Interrompe a iteração assim que a política é encontrada
			}
		}
		return true // Continua a iteração para a próxima página
	})

	if err != nil {
		return "", err
	}

	if policyArn == "" {
		return "", fmt.Errorf("policy not found: %s", policyName)
	}

	return policyArn, nil
}
