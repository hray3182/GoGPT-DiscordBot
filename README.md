# GoGPT-DiscordBot

GoGPT-DiscordBot is a minimal implementation of a Discord bot powered by OpenAI's GPT models.

## How to Use:

### Running from Source Code:
Execute the following command:
```bash
go run main.go
```

### Running from Binary:
Download the binary release and double-click to run.

### Environment Configuration:

Create a `.env` file with the following content:
```bash
DISCORD_BOT_TOKEN="YOUR_DISCORD_BOT_TOKEN"
OPENAI_API_KEY="YOUR_OPENAI_API_KEY"
```

## Configuration:

Commands can be added or modified by editing the `chat.json` file. The fields to be included are:

- `command`: The actual command (required).
- `description`: Description of the command (required).
- `parameter_name`: Name of the parameter (required).
- `parameter_description`: Description of the parameter (required).
- `openai_model`: The OpenAI model to use (optional, default: gpt-3.5-turbo).
- `prompt`: The prompt for OpenAI (optional).

The server will concatenate the `prompt` with the `parameter` before sending the request to the model.

**Example**:
```go
prompt := "Help me translate the following into English: \n"
parameter := "你好?"

// Message sent to GPT:
//Help me translate the following into English: 
//你好
```

For a list of all supported models, refer to:
- [GPT-3.5 Models](https://platform.openai.com/docs/models/gpt-3-5)
- [GPT-4 Models](https://platform.openai.com/docs/models/gpt-4)

## Notes:

1. This project is a minimalistic implementation for a Discord bot leveraging GPT.
2. Discord has a message limit of 2000 characters, and this app not handle that.
3. This project doesn't account for Discord API rate limits. Excessive use on large servers might exceed these limits.
4. Similarly, this project doesn't account for OpenAI API rate limits.
5. Dependencies include the `go-openai` and `discordgo` packages.

