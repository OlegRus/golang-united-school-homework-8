package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

const FILE_PERMISSION = 0644

const (
	OPERATION_FLAG = "operation"
	ITEM_FLAG      = "item"
	ID_FLAG        = "id"
	FILE_NAME_FLAG = "fileName"
)

const (
	ADD        = "add"
	LIST       = "list"
	FIND_BY_ID = "findById"
	REMOVE     = "remove"
)

var (
	ErrFileName  = fmt.Errorf("-fileName flag has to be specified")
	ErrOperation = fmt.Errorf("-operation flag has to be specified")
	ErrIdFlag    = fmt.Errorf("-id flag has to be specified")
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func parseArgs() Arguments {
	operationVar := flag.String(OPERATION_FLAG, "", "")
	itemVar := flag.String(ITEM_FLAG, "", "")
	idVar := flag.String(ID_FLAG, "", "")
	fileNameVar := flag.String(FILE_NAME_FLAG, "", "")

	return Arguments{
		OPERATION_FLAG: *operationVar,
		ITEM_FLAG:      *itemVar,
		ID_FLAG:        *idVar,
		FILE_NAME_FLAG: *fileNameVar,
	}
}

func readUsersFromJsonFile(reader io.Reader) ([]User, error) {
	byteBuffer, err := ioutil.ReadAll(reader)
	fmt.Print(string(byteBuffer))
	if err != nil {
		return nil, err
	}

	users := make([]User, 0)
	if len(byteBuffer) == 0 {
		return users, nil
	}

	if err = json.Unmarshal(byteBuffer, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func writeUsersToJsonFile(users *[]User, writer io.Writer) error {
	byteBuffer, err := json.Marshal(users)
	if err != nil {
		return err
	}

	if _, err := writer.Write(byteBuffer); err != nil {
		return err
	}
	return nil
}

func list(fileName string, writer io.Writer) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, FILE_PERMISSION)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		buffer := [2048]byte{}
		n, err := file.Read(buffer[:])

		writer.Write(buffer[:n])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func add(fileName, item string, writer io.Writer) error {
	if item == "" {
		return fmt.Errorf("-item flag has to be specified")
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, FILE_PERMISSION)
	if err != nil {
		return err
	}
	defer file.Close()

	users, err := readUsersFromJsonFile(file)
	if err != nil {
		return err
	}

	addingUser := User{}
	if err = json.Unmarshal([]byte(item), &addingUser); err != nil {
		return err
	}

	for _, existedUser := range users {
		if existedUser.Id == addingUser.Id {
			if _, err = fmt.Fprintf(writer, "Item with id %s already exists", existedUser.Id); err != nil {
				return err
			}
		}
	}

	users = append(users, addingUser)
	if err = writeUsersToJsonFile(&users, file); err != nil {
		return err
	}

	return nil
}

func findById(id, fileName string, writer io.Writer) error {
	if id == "" {
		return ErrIdFlag
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, FILE_PERMISSION)
	if err != nil {
		return err
	}
	defer file.Close()

	users, err := readUsersFromJsonFile(file)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Id == id {
			byteBuffer, err := json.Marshal(user)
			if err != nil {
				return err
			}
			if _, err = writer.Write(byteBuffer); err != nil {
				return err
			}
		}
	}
	_, err = writer.Write([]byte(""))
	return err
}

func remove(id, fileName string, writer io.Writer) error {
	if id == "" {
		return ErrIdFlag
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	users, err := readUsersFromJsonFile(file)
	if err != nil {
		return err
	}

	for i, user := range users {
		if user.Id == id {
			newUsers := append(users[:i], users[i+1:]...)
			if err = writeUsersToJsonFile(&newUsers, file); err != nil {
				return err
			}
			return nil
		}
	}

	_, err = fmt.Fprintf(writer, "Item with id %s not found", id)
	return err
}

func Perform(args Arguments, writer io.Writer) error {
	fileName := args[FILE_NAME_FLAG]
	operation := args[OPERATION_FLAG]
	item := args[ITEM_FLAG]
	id := args[ID_FLAG]
	if fileName == "" {
		return ErrFileName
	}
	if operation == "" {
		return ErrOperation
	}
	switch operation {
	case ADD:
		return add(fileName, item, writer)
	case LIST:
		return list(fileName, writer)
	case FIND_BY_ID:
		return findById(id, fileName, writer)
	case REMOVE:
		return remove(id, fileName, writer)
	default:
		return fmt.Errorf("Operation %s not allowed!", operation)
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
