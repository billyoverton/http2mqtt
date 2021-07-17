package main

import (
	"fmt"
  "log"
  "io"
  "os"
  "os/signal"
  "net/http"

  "github.com/spf13/cobra"
  "github.com/spf13/viper"
  "github.com/mitchellh/go-homedir"
  MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTMessage struct {
  topic string
  message string
}

var (
  cfgFile string

  rootCmd = &cobra.Command{
    Use: "http2mqtt",
    Short: "A HTTP to MQTT Broker Bridge",
    Long: "A HTTP to MQTT Broker Bridge",
    Run: root,
  }

  mqttClient MQTT.Client

  messageChan = make(chan *MQTTMessage, 10)
)

func init() {
  cobra.OnInitialize(initConfig)

  rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.http2mqtt.yaml)")
  rootCmd.PersistentFlags().IntP("port", "p", 8080, "port to listen for web connections")
  rootCmd.PersistentFlags().StringP("broker", "b", "localhost", "MQTT broker address")
  rootCmd.PersistentFlags().IntP("brokerport", "q", 1883, "port to listen for web connections")
  rootCmd.PersistentFlags().StringP("username", "u", "", "broker username")
  rootCmd.PersistentFlags().StringP("password", "P", "", "broker password")

  viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
  viper.BindPFlag("broker", rootCmd.PersistentFlags().Lookup("broker"))
  viper.BindPFlag("brokerport", rootCmd.PersistentFlags().Lookup("brokerport"))
  viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
  viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))

}

func initConfig() {
  if cfgFile != "" {
    viper.SetConfigFile(cfgFile)
  } else {
    home, err := homedir.Dir()
    cobra.CheckErr(err)

    viper.AddConfigPath(home)
    viper.SetConfigName(".http2mqtt")
  }

  viper.AutomaticEnv()

  if err := viper.ReadInConfig(); err == nil {
    fmt.Println("Using config file: ", viper.ConfigFileUsed())
  }
}

func urlHandler(w http.ResponseWriter, r *http.Request) {
  topic := r.URL.Path[1:]

  log.Printf("Topic: %s", topic)

  switch r.Method {
    case "GET":
      messages := r.URL.Query()["message"]

      if len(messages) > 0 {
        for _, message := range messages {
          messageChan <- &MQTTMessage {
            message: message,
            topic: topic,
          }

        }
      }
    case "POST":
      body, err := io.ReadAll(r.Body)

      if err == nil {
        messageChan <- &MQTTMessage {
          message: string(body),
          topic: topic,
        }
      } else {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("Unable to parse message body"))
      }
    default:
      w.WriteHeader(http.StatusMethodNotAllowed)
      w.Write([]byte("Only GET and POST methods are allowed."))
  }
}

func mqttConnectHandler(client MQTT.Client) {
  log.Printf("Connected to broker")
}

func mqttConnectionLostHandler(client MQTT.Client, err error) {
  log.Printf("Connection to broker lost: %v", err)
}

func mqttReconnectingHandler(client MQTT.Client, opts *MQTT.ClientOptions) {
  log.Printf("Attempting to reconnect to broker")
}

func root(c *cobra.Command, args []string) {
  broker := fmt.Sprintf("tcp://%s:%d", viper.GetString("broker"), viper.GetInt("brokerport"))
  username := viper.GetString("username")
  password := viper.GetString("password")

  opts := MQTT.NewClientOptions()
  opts.AddBroker(broker)

  if username != "" {
    opts.SetUsername(username)
    opts.SetPassword(password)
  }

  opts.OnConnect = mqttConnectHandler
  opts.OnConnectionLost = mqttConnectionLostHandler
  opts.OnReconnecting = mqttReconnectingHandler
  opts.AutoReconnect = true


  mqttClient := MQTT.NewClient(opts)
  if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
        log.Fatalf("MQQT.Connect(): %v", token.Error())
  }

  go func() {
    for message := range messageChan {
      log.Printf("Sending message %v to topic %v", message.message, message.topic)
      mqttClient.Publish(message.topic, 0, false, message.message)
    }
  }()



  addr := fmt.Sprintf(":%d", viper.GetInt("port"))

  go func() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", urlHandler)

    server := &http.Server{
      Addr: addr,
      Handler: mux,
    }

    if err := server.ListenAndServe(); err != http.ErrServerClosed {
      log.Fatalf("ListenAndServe(): %v", err)
    }
  }()

  log.Printf("Listening on %v", addr)

  // Wait for the interupt signal
  // Setup signal capture
  stop := make(chan os.Signal, 1)
  signal.Notify(stop, os.Interrupt)
  <-stop

  close(messageChan)

  log.Printf("Finished")
}

func main() {
  rootCmd.Execute()
}
