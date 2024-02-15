package add_groups_list

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Função para listar grupos IAM existentes e retornar o grupo selecionado
func ListGroups(sess *session.Session) *iam.Group {
	svc := iam.New(sess)

	result, err := svc.ListGroups(nil)
	if err != nil {
		log.Fatal("Erro ao listar grupos:", err)
	}

	groups := result.Groups
	fmt.Println("Grupos existentes:")
	for i, group := range groups {
		fmt.Printf("%d. %s\n", i+1, *group.GroupName)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Escolha o número do grupo a ser adicionado ao usuário: ")
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(groups) {
		fmt.Println("Escolha inválida!")
		os.Exit(1)
	}

	return groups[choice-1]
}

// Função para adicionar usuário ao grupo selecionado
func AddUserToGroup(sess *session.Session, userName string, group *iam.Group) {
	svc := iam.New(sess)

	_, err := svc.AddUserToGroup(&iam.AddUserToGroupInput{
		GroupName: aws.String(*group.GroupName),
		UserName:  aws.String(userName),
	})
	if err != nil {
		log.Fatal("Erro ao adicionar usuário ao grupo:", err)
	}
	fmt.Printf("Usuário %s adicionado ao grupo %s com sucesso.\n", userName, *group.GroupName)
}
