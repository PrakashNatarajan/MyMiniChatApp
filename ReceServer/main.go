//Source ==> https://tutorialedge.net/projects/chat-system-in-go-and-react/part-4-handling-multiple-clients/

package main

import (
    "fmt"
    "log"
    "regexp"
    "context"
    "net/http"
    "crypto/rand"
    "encoding/json"
    "encoding/base64"
    "github.com/gorilla/websocket"
    amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
  ID   string
  Name string
  Conn *websocket.Conn
  Pool *Pool
}

type Pool struct {
  Register   chan *Client
  Unregister chan *Client
  Clients    map[*Client]bool
  Broadcast  chan Message
}

func NewPool() *Pool {
  return &Pool{
    Register:   make(chan *Client),
    Unregister: make(chan *Client),
    Clients:    make(map[*Client]bool),
    Broadcast:  make(chan Message),
  }
}

var upgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
  WriteBufferSize: 1024,
  CheckOrigin: func(req *http.Request) bool { return true },
}

func connUpgrade(resWtr http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
  conn, err := upgrader.Upgrade(resWtr, req, nil)
  if err != nil {
    log.Println(err)
    return nil, err
  }
  return conn, nil
}

type Message struct {
  Type string `json:"type"`
  Sender string `json:"sender"`
  Recipient string `json:"recipient"`
  Text string `json:"text"`
}

type ClientConnMsg struct {
  Type string `json:"type"`
  Body string `json:"body"`
}

func failOnError(err error, msg string) {
  if err != nil {
    log.Panicf("%s: %s", msg, err)
  }
}

var channel *amqp.Channel
var queue amqp.Queue

func (pool *Pool)EventQueueConn() () {
  conn, err := amqp.Dial("amqp://MsgQueAdmin:MyQueueS123@172.17.0.2:5672/")
  failOnError(err, "Failed to connect to RabbitMQ")

  channel, err = conn.Channel()
  failOnError(err, "Failed to open a channel")

  queue, err = channel.QueueDeclare(
    "TestQueue", // name
    false,   // durable
    false,   // delete when unused
    false,   // exclusive
    false,   // no-wait
    nil,     // arguments
  )
  failOnError(err, "Failed to declare a queue")
}

func (clnt *Client) ReceivingHandler() {
  defer func() {
    clnt.Pool.Unregister <- clnt
    clnt.Conn.Close()
  }()

  for {
    message := &Message{}
    err := clnt.Conn.ReadJSON(message)
    if err != nil {
      log.Println(err)
      return
    }

    clnt.SendMessageQueue(message)
    fmt.Printf("Message Received: %+v\n", message)
  }
}

func RandomStringGenerate(count int) (string) {
  var ranStr string
  bt := make([]byte, count)
  _, err := rand.Read(bt)
  if err != nil {
    fmt.Println("error:", err)
    return "error"
  }
  ranStr = base64.URLEncoding.EncodeToString(bt)
  ranStr = regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(ranStr, "")
  return ranStr
}


func (pool *Pool)RegisterClient(client *Client) {
  pool.Clients[client] = true
  fmt.Println("Size of Connection Pool: ", len(pool.Clients))
  for client, _ := range pool.Clients {
    fmt.Println(client)
    client.Conn.WriteJSON(ClientConnMsg{Type: "ClientConn", Body: "New User Joined..."})
  }
}

func (pool *Pool)UnRegisterClient(client *Client) {
  delete(pool.Clients, client)
  fmt.Println("Size of Connection Pool: ", len(pool.Clients))
  for client, _ := range pool.Clients {
    client.Conn.WriteJSON(ClientConnMsg{Type: "ClientConn", Body: "User Disconnected..."})
  }
}

func (pool *Pool)ReciveSendClient(message Message) {
  fmt.Println("Sending message to right clients in Pool")
  for client, _ := range pool.Clients {
    if message.Sender == client.Name {
      if err := client.Conn.WriteJSON(message); err != nil {
        fmt.Println(err)
        return
      }
    }
    if message.Recipient == client.Name {
      if err := client.Conn.WriteJSON(message); err != nil {
        fmt.Println(err)
        return
      }
    }
  }
}

func (pool *Pool) ManageClientConns() {
  for {
    select {
    case client := <-pool.Register:
      pool.RegisterClient(client)
      break
    case client := <-pool.Unregister:
      pool.UnRegisterClient(client)
      break
    case message := <-pool.Broadcast:
      pool.ReciveSendClient(message)
    }
  }
}

func (client *Client)SendMessageQueue(message *Message) {
  strMsg, err := json.Marshal(message)
  if err != nil {
    panic (err)
  }
  log.SetFlags(log.LstdFlags | log.Lmicroseconds) //To Print in Millisecond
  log.Printf("Sending a message: %s", message)
  err = channel.PublishWithContext(context.Background(),
    "",     // exchange
    queue.Name, // routing key
    false,  // mandatory
    false,  // immediate
    amqp.Publishing{
      ContentType: "text/plain",
      Body:        []byte(strMsg),
    })
  failOnError(err, "Failed to receive messages from queue")
}

func serveWs(pool *Pool, resWtr http.ResponseWriter, req *http.Request) {
  fmt.Println("WebSocket Endpoint Hit")
  conn, err := connUpgrade(resWtr, req)
  if err != nil {
    fmt.Fprintf(resWtr, "%+v\n", err)
  }

  xkey := req.Header.Get("Sec-WebSocket-Key")
  fmt.Println(xkey)

  CurrUser := req.FormValue("CurrUser")

  fmt.Println(CurrUser)

  client := &Client{
	  Name: CurrUser,
    Conn: conn,
    Pool: pool,
  }

  pool.Register <- client
  client.ReceivingHandler()
}

func setupRoutes() {
  pool := NewPool()
  pool.EventQueueConn()
  go pool.ManageClientConns()
/*
  http.HandleFunc("/", func(resWtr http.ResponseWriter, req *http.Request) {
    http.ServeFile(resWtr, req, "WebClient11.html")
  })
*/
  http.HandleFunc("/sender", func(resWtr http.ResponseWriter, req *http.Request) {
    serveWs(pool, resWtr, req)
  })
}

func main() {
  fmt.Println("Distributed Chat App v0.01")
  setupRoutes()
  http.ListenAndServe(":5600", nil)
}