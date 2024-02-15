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

// ListGroups lista os grupos IAM existentes e retorna os grupos selecionados
func ListGroups(sess *session.Session) []*iam.Group {
	svc := iam.New(sess)

	result, err := svc.ListGroups(nil)
	if err != nil {
		log.Fatalf("Erro ao listar grupos: %v", err)
	}

	groups := result.Groups
	fmt.Println("Grupos existentes:")
	for i, group := range groups {
		fmt.Printf("[%d]. %s\n", i+1, *group.GroupName)
	}

	return groups
}

// AddUserToGroups adiciona usuário ao(s) grupo(s) selecionado(s)
func AddUserToGroups(sess *session.Session, userName string, groups []*iam.Group) {
	svc := iam.New(sess)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Deseja adicionar o usuário em algum grupo? \n[1] SIM\n[2] NÃO \n:")
		choice, err := getUserChoice(reader)
		if err != nil {
			log.Fatalf("Erro ao ler a escolha do usuário: %v", err)
		}

		if choice == 2 {
			break
		}

		fmt.Println("Escolha o número do grupo para adicionar o usuário (ou 0 para sair):")
		for i, group := range groups {
			fmt.Printf("[%d]. %s\n", i+1, *group.GroupName)
		}
		fmt.Printf(":")

		choice, err = getUserChoice(reader)
		if err != nil {
			log.Fatalf("Erro ao ler a escolha do usuário: %v", err)
		}

		if choice == 0 {
			break
		}

		if choice < 0 || choice > len(groups) {
			fmt.Println("Escolha inválida!")
			continue
		}

		group := groups[choice-1]
		_, err = svc.AddUserToGroup(&iam.AddUserToGroupInput{
			GroupName: aws.String(*group.GroupName),
			UserName:  aws.String(userName),
		})
		if err != nil {
			log.Fatalf("Erro ao adicionar usuário ao grupo %s: %v", *group.GroupName, err)
		}
		fmt.Printf("Usuário %s adicionado ao grupo %s com sucesso.\n", userName, *group.GroupName)
	}
}

// getUserChoice lê a escolha do usuário e retorna um número inteiro
func getUserChoice(reader *bufio.Reader) (int, error) {
	choiceStr, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	choiceStr = strings.TrimSpace(choiceStr)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil {
		return 0, err
	}
	return choice, nil
}
