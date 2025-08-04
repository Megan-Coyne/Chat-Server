package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
    "os"
    "strings"
    "sync"
    // "slices"
)

var broadcast = make(chan string)

type Credentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

const BUFFERSIZE = 1024

var (
    allClients  = make(map[net.Conn]string)
    newClient   = make(chan net.Conn)
    lostClient  = make(chan net.Conn)
    authClients = make(map[net.Conn]bool)
    mutex       = &sync.Mutex{}
    displayedList []string
)

func main() {
    if len(os.Args) != 2 {
        fmt.Printf("Usage: %s <port>\n", os.Args[0])
        os.Exit(0)
    }
    port := os.Args[1]
    server, err := net.Listen("tcp", ":"+port)
    if err != nil {
        fmt.Printf("Cannot listen on port '%s'!\n", port)
        os.Exit(2)
    }
    fmt.Println("EchoServer in GoLang developed by Phu Phung, revised by Megan Coyne")
    fmt.Printf("Server is listening on port '%s' ...\n", port)

    go func() {
        for {
            clientConn, _ := server.Accept()
            newClient <- clientConn
        }
    }()

    for {
        select {
        case clientConn := <-newClient:
            mutex.Lock()
            allClients[clientConn] = clientConn.RemoteAddr().String()
            mutex.Unlock()
            go clientHandler(clientConn)

        case clientConn := <-lostClient:
            mutex.Lock()
            delete(allClients, clientConn)
            delete(authClients, clientConn)
            mutex.Unlock()
            clientConn.Close()
        }
    }
}

var loggedOut = false

func clientHandler(clientConn net.Conn) {
    defer func() {
        // Only remove the user from the list when they explicitly log out using `.exit`
        if !loggedOut {
            mutex.Lock()
            delete(allClients, clientConn)
            delete(authClients, clientConn)
            mutex.Unlock()
        }
    }()

    scanner := bufio.NewScanner(clientConn)
    loggedOut := false

    for scanner.Scan() {
        message := scanner.Text()
        parts := strings.Fields(message)

        if len(parts) == 3 && parts[0] == "login" {
            login(clientConn, parts[1], parts[2])
        } else if len(parts) > 1 && parts[0] == ">" {
            if isAuthenticated(clientConn) {
                content := strings.Join(parts[1:], " ")
                broadcastMessages(content, clientConn)
            } else {
                clientConn.Write([]byte("You must be logged in to send messages.\n"))
            }
        } else if strings.HasPrefix(parts[0], "[To:") && strings.HasSuffix(parts[0], "]") {
            if isAuthenticated(clientConn) {
                recipientUsername := strings.TrimPrefix(parts[0], "[To:")
                recipientUsername = strings.TrimSuffix(recipientUsername, "]")
                sendPrivateMessage(recipientUsername, message, clientConn)
            } else {
                clientConn.Write([]byte("You must be logged in to send private messages.\n"))
            }
        } else if len(parts) == 1 && parts[0] == ".userlist" {
            notifyUserList()
        } else if len(parts) == 1 && parts[0] == ".exit" {
            if isAuthenticated(clientConn) {
                logout(clientConn)
                loggedOut = true
                break
            } else {
                clientConn.Write([]byte("You can't logout if your're not logged in.\n"))
            }
        } else {
            clientConn.Write([]byte("Invalid command. Use 'login <username> <password>' to authenticate.\n"))
        }
    }

    // If the client didn't log out explicitly (i.e., closed the terminal), don't remove them from the list
    if loggedOut {
        clientConn.Close()
    }
}

func login(clientConn net.Conn, username, password string) {
    if checkLogin(username, password) {
        if(stringInArray(displayedList, username)){
            mutex.Lock()
            authClients[clientConn] = true
            allClients[clientConn] = username
            mutex.Unlock()
            clientConn.Write([]byte("You are already logged in!"))
        } else {
            mutex.Lock()
            authClients[clientConn] = true
            allClients[clientConn] = username
            mutex.Unlock()
            displayedList = append(displayedList, username)
            clientConn.Write([]byte("You are authenticated! Welcome to the chat system!\n" + 
            "Type '[To:Receiver] Message' to send to a specific user.\n" +
            "Type .userlist to request latest online users.\n"+
            "Type .exit to logout and close the connection."))
    
            ipAddress := clientConn.RemoteAddr().String()
            newUserMessage := fmt.Sprintf("New user '%s' logged into Chat System from %s\n", username, ipAddress)
            broadcastMessages(newUserMessage, clientConn)
            notifyUserList()
        }
    } else {
        clientConn.Write([]byte("Authentication failed!\n"))
    }
}

func checkLogin(username, password string) bool {
    ok, _ := checkAccount("./credentials.json", username, password)
    // displayedList = append(displayedList, username)
    // fmt.Println(displayedList)

    return ok
}

func checkAccount(filename, username, password string) (bool, error) {
    file, err := os.ReadFile(filename)
    if err != nil {
        return false, err
    }

    var creds []Credentials
    if err := json.Unmarshal(file, &creds); err != nil {
        return false, err
    }

    for _, cred := range creds {
        if cred.Username == username && cred.Password == password {
            return true, nil
        }
    }
    return false, nil
}

func isAuthenticated(clientConn net.Conn) bool {
    mutex.Lock()
    defer mutex.Unlock()
    return authClients[clientConn]
}

func broadcastMessages(msg string, sender net.Conn) {
    mutex.Lock()
    defer mutex.Unlock()

    fmt.Println("Broadcasting: " + msg + "\n")

    senderName := allClients[sender]

    for client := range authClients {
        client.Write([]byte(fmt.Sprintf("Public message from '%s': %s\n", senderName, msg)))
    }
}

func notifyUserList() {
    fmt.Println("Printing userlist: " + strings.Join(displayedList, ", ") + "\n")
    mutex.Lock()
    defer mutex.Unlock()

    userListMessage1 := "Online users: " + strings.Join(displayedList, ", ") + "\n"
    if len(displayedList) == 1{
        userListMessage1 = "Online users: " + displayedList[0] + "\n"
    }

    for client := range authClients {
        client.Write([]byte(userListMessage1))
        client.Write([]byte(fmt.Sprintf("# of connected clients: %d\n\n", len(displayedList))))
    }
}

func sendPrivateMessage(username, message string, sender net.Conn) {
    mutex.Lock()
    defer mutex.Unlock()

    var recipientConn net.Conn
    var found bool
    parts := strings.Fields(message)
    messageContent := strings.Join(parts[1:], " ")

    for conn, user := range allClients {
        if user == username {
            recipientConn = conn
            found = true
            break
        }
    }

    if found {
        recipientConn.Write([]byte(fmt.Sprintf("Private message from '%s': %s\n", allClients[sender], messageContent)))
    } else {
        sender.Write([]byte(fmt.Sprintf("User '%s' not found or not logged in.\n", username)))
    }
}

func logout(clientConn net.Conn) {
    mutex.Lock()

    username := allClients[clientConn]

    displayedList = deleteString(displayedList, username)
    fmt.Println(displayedList)
    logoutMessage := fmt.Sprintf("User '%s' has logged out.\n", username)
    
    for client := range authClients {
        client.Write([]byte(logoutMessage))
    }

    delete(authClients, clientConn)
    delete(allClients, clientConn)

    mutex.Unlock()

    // Also remove the username from the displayed list
    displayedList = deleteString(displayedList, username)
    fmt.Println(displayedList)

    clientConn.Write([]byte(fmt.Sprintf("You have been logged out, %s.\n", username)))
    clientConn.Close()
}



func deleteString(slice []string, target string) []string {
    newSlice := make([]string, 0, len(slice))
    for _, str := range slice {
        if str != target {
            newSlice = append(newSlice, str)
        }
    }

    return newSlice
}

func stringInArray(arr []string, str string) bool {
	for _, element := range arr {
		if element == str {
			return true
		}
	}
	return false
}