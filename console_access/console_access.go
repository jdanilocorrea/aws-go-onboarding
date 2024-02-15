package console_access

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// CreateConsoleAccess cria acesso ao console AWS para o novo usuário
func CreateConsoleAccess(sess *session.Session, profile, userName string) {
	// Crie um cliente IAM a partir da sessão
	svc := iam.New(sess)

	// Gere uma senha aleatória temporária
	tempPassword, err := GenerateRandomPassword()
	if err != nil {
		log.Fatalf("Erro ao gerar a senha temporária: %v", err)
	}

	// Crie um perfil de login para o usuário com a senha temporária
	_, err = svc.CreateLoginProfile(&iam.CreateLoginProfileInput{
		UserName:              aws.String(userName),
		Password:              aws.String(tempPassword),
		PasswordResetRequired: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("Erro ao criar acesso ao console da AWS: %v", err)
	}

	// // Verifique se o resultado e o perfil de login não são nulos
	// if result != nil && result.LoginProfile != nil {
	// 	fmt.Println("Perfil de login criado com sucesso:")
	// 	fmt.Println("User name:", *result.LoginProfile.UserName)
	// 	fmt.Println("Password reset required:", *result.LoginProfile.PasswordResetRequired)
	// } else {
	// 	fmt.Println("Erro: O objeto result ou result.LoginProfile é nil.")
	// }

	// Crie o conteúdo do arquivo console-access
	consoleAccessInfo := fmt.Sprintf("Link: [%s]\nUser: %s\nPassword: %s\n", profile, userName, tempPassword)

	// Abrir o arquivo no modo de anexo para escrita
	file, err := os.OpenFile("console-access", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo 'console-access': %v", err)
	}
	defer file.Close()

	// Escrever no arquivo
	_, err = file.WriteString(consoleAccessInfo)
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo 'console-access': %v", err)
	}

	fmt.Println("Acesso ao console da AWS criado com sucesso para o usuário:", userName)
	fmt.Println("Arquivo 'console-access' criado com sucesso.")
}

// GenerateRandomPassword gera uma senha aleatória segura
func GenerateRandomPassword() (string, error) {
	// Defina o tamanho da senha
	const passwordLength = 12

	// Crie um slice de bytes para armazenar a senha
	password := make([]byte, passwordLength)

	// Use o tempo atual como semente para a função de geração de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Gere números aleatórios para cada caractere da senha
	for i := 0; i < passwordLength; i++ {
		// Gere um número aleatório entre 33 e 126, que corresponde aos caracteres imprimíveis ASCII
		password[i] = byte(rand.Intn(94) + 33)
	}

	// Converta os bytes em uma string base64
	return base64.StdEncoding.EncodeToString(password), nil
}
