package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type chatSetting struct {
	Command              string `json:"command"`
	Description          string `json:"description"`
	ParameterName        string `json:"parameter_name"`
	ParameterDescription string `json:"parameter_description"`
	OpenAIModel          string `json:"openai_model"`
	Prompt               string `json:"prompt"`
}

type settingsWrapper struct {
	Commands []chatSetting `json:"commands"`
}

var OpenAIClient *openai.Client

func init() {

	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	APIKey := os.Getenv("OPENAI_API_KEY")

	if APIKey == "" {
		log.Fatal("OpenAI APIKey is empty, please check your .env file")
	}
	OpenAIClient = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	if OpenAIClient == nil {
		log.Fatal("create openAI client fail")
	}
}

func main() {
	var chatSettings settingsWrapper

	jsonFile, err := os.ReadFile("chat.json")
	if err != nil {
		log.Fatal("Read chat.json fail: ", err)
	}

	err = json.Unmarshal(jsonFile, &chatSettings)
	if err != nil {
		log.Fatal(err)
	}

	// check chat settings
	for i, c := range chatSettings.Commands {
		if c.OpenAIModel == "" {
			chatSettings.Commands[i].OpenAIModel = "gpt-3.5-turbo"
		}
	}

	discordToken := os.Getenv("DISCORD_BOT_TOKEN")
	if discordToken == "" {
		log.Fatal("Discord token is empty, please check your .env file")
	}

	log.Println("Create discord bot session")
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal(err)
	}

	defer dg.Close()

	commandMap := make(map[string]*discordgo.ApplicationCommand)

	log.Println("Create command")
	// create command
	for _, c := range chatSettings.Commands {
		commandMap[c.Command] = c.createCommand()
	}
	log.Println("add handler")
	//add handler
	for _, c := range chatSettings.Commands {
		dg.AddHandler(c.handler)
	}

	log.Println("connect to discord")
	err = dg.Open()
	if err != nil {
		log.Fatal("can not connect to discord: ", err)
	}

	log.Println("apply command")
	// apply comamnd
	for _, c := range chatSettings.Commands {
		_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", commandMap[c.Command])
		if err != nil {
			log.Printf("can not create %v command: %v", c.Command, err)
		}
	}

	log.Println("Bot is running. Press Ctrl-C to exit.")
	select {}
}

func (c chatSetting) handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if i.ApplicationCommandData().Name != c.Command {
		return
	}
	log.Printf("Receive Command %v\n", c.Command)

	message := i.ApplicationCommandData().Options[0].StringValue()
	content := fmt.Sprintf("%v: %v", c.Command, message)
	response := &discordgo.WebhookEdit{
		Content: &content,
	}
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource}); err != nil {
		log.Println("Cannot send deferred message:", err)
		return
	}

	go func() {
		stream := c.createStream(message)
		if stream == nil {
			log.Println("stream is nil")
		}
		defer stream.Close()
		answer := ""
		buffer := make(chan string, 100)

		go func() {
			for {
				res, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					close(buffer)
					return
				}

				if err != nil {
					fmt.Printf("\nStream error: %v\n", err)
					return
				}

				fmt.Printf(res.Choices[0].Delta.Content)
				buffer <- res.Choices[0].Delta.Content
			}
		}()

		for {
			select {
			case data, ok := <-buffer:
				if !ok {
					msg, err := s.InteractionResponseEdit(i.Interaction, response)
					if err != nil {
						log.Println("Cannot send interaction response:", err)
					}
					err = s.MessageReactionAdd(i.ChannelID, msg.ID, "✅")
					if err != nil {
						log.Println("Cannot add reaction:", err)
					}

					return
				}
				answer += data
				response.Content = &answer
			default:
				if _, err := s.InteractionResponseEdit(i.Interaction, response); err != nil {
					log.Println("Cannot send interaction response:", err)
				}
			}
		}
	}()
}

func (c chatSetting) createStream(content string) *openai.ChatCompletionStream {
	resp, err := OpenAIClient.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: c.OpenAIModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("%s\ninput：%s, output:", c.Prompt, content),
				},
			},
			Stream: true,
		},
	)
	if err != nil {
		log.Println("Cannot create chat completion:", err)
		return nil
	}

	// check the response status code
	if resp.GetResponse().StatusCode != 200 {
		log.Println("Create chat completion failed, status code:", resp.GetResponse().Status)
		return nil
	}

	return resp
}

func (c chatSetting) createCommand() *discordgo.ApplicationCommand {

	cmd := &discordgo.ApplicationCommand{
		Name:        c.Command,
		Description: c.Description,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        c.ParameterName,
				Description: c.ParameterDescription,
				Required:    true,
			},
		},
	}

	return cmd
}
