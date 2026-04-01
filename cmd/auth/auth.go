package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/config"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
	"github.com/crowdy/lm-cli/internal/prompt"
)

// Cmd is the auth command group.
var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

func init() {
	loginCmd.Flags().StringP("type", "t", "longterm", "token type: longterm, stateless, v2")
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(switchCmd)
	tokenCmd.Flags().Bool("check", false, "exit 0 if token is valid, exit 2 if not (no output)")
	Cmd.AddCommand(tokenCmd)
	Cmd.AddCommand(removeCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure authentication for a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenType, _ := cmd.Flags().GetString("type")
		profileName := getProfileFlag(cmd)

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		switch tokenType {
		case "longterm":
			return loginLongterm(profileName, cfg)
		case "stateless":
			return loginStateless(profileName, cfg)
		case "v2":
			return loginV2(profileName, cfg)
		default:
			return &lmerrors.ValidationError{
				Field:   "type",
				Message: fmt.Sprintf("unknown token type %q (use: longterm, stateless, v2)", tokenType),
			}
		}
	},
}

func loginLongterm(profileName string, cfg *config.Config) error {
	var err error

	channelID := config.EnvOr(config.EnvChannelID, "")
	if channelID == "" {
		channelID, err = prompt.String("Channel ID")
		if err != nil {
			return err
		}
	}

	token, err := prompt.Password("Long-term Channel Access Token")
	if err != nil {
		return err
	}

	// Save profile
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]config.Profile{}
	}
	cfg.Profiles[profileName] = config.Profile{
		ChannelID: channelID,
		TokenType: "longterm",
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = profileName
	}
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Save token (no expiry for longterm)
	tokens, err := config.LoadTokens()
	if err != nil {
		return err
	}
	tokens.Set(profileName, config.TokenEntry{
		Token:     token,
		TokenType: "longterm",
	})
	if err := tokens.Save(); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Logged in to profile %q with longterm token\n", profileName)
	verifyToken(token)
	return nil
}

func loginStateless(profileName string, cfg *config.Config) error {
	var err error

	channelID := config.EnvOr(config.EnvChannelID, "")
	if channelID == "" {
		channelID, err = prompt.String("Channel ID")
		if err != nil {
			return err
		}
	}

	secret := config.EnvOr(config.EnvSecret, "")
	if secret == "" {
		secret, err = prompt.Password("Channel Secret")
		if err != nil {
			return err
		}
	}

	// Issue token immediately to verify credentials
	fmt.Fprintln(os.Stderr, "Issuing stateless token...")
	token, expiresAt, err := api.IssueStatelessToken(channelID, secret)
	if err != nil {
		return err
	}

	// Save profile
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]config.Profile{}
	}
	cfg.Profiles[profileName] = config.Profile{
		ChannelID: channelID,
		TokenType: "stateless",
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = profileName
	}
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Save credentials
	creds, err := config.LoadCredentials()
	if err != nil {
		return err
	}
	creds.Set(profileName, config.Credentials{ChannelSecret: secret})
	if err := creds.Save(); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	// Save token
	tokens, err := config.LoadTokens()
	if err != nil {
		return err
	}
	tokens.Set(profileName, config.TokenEntry{
		Token:     token,
		ExpiresAt: expiresAt,
		TokenType: "stateless",
	})
	if err := tokens.Save(); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	jst := time.FixedZone("JST", 9*60*60)
	fmt.Fprintf(os.Stderr, "Logged in to profile %q (token expires %s / %s JST)\n",
		profileName,
		expiresAt.Format(time.RFC3339),
		expiresAt.In(jst).Format("2006-01-02 15:04"))
	verifyToken(token)
	return nil
}

func loginV2(profileName string, cfg *config.Config) error {
	var err error

	channelID := config.EnvOr(config.EnvChannelID, "")
	if channelID == "" {
		channelID, err = prompt.String("Channel ID")
		if err != nil {
			return err
		}
	}

	privateKeyFile, err := prompt.String("Private Key file path")
	if err != nil {
		return err
	}

	// Expand ~ and resolve to absolute path
	if strings.HasPrefix(privateKeyFile, "~/") {
		home, _ := os.UserHomeDir()
		privateKeyFile = filepath.Join(home, privateKeyFile[2:])
	}
	privateKeyFile, err = filepath.Abs(privateKeyFile)
	if err != nil {
		return fmt.Errorf("resolving key path: %w", err)
	}

	// Issue token immediately to verify credentials
	fmt.Fprintln(os.Stderr, "Issuing v2 token via JWT assertion...")
	token, expiresAt, err := api.IssueV2Token(channelID, privateKeyFile)
	if err != nil {
		return err
	}

	// Save profile
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]config.Profile{}
	}
	cfg.Profiles[profileName] = config.Profile{
		ChannelID: channelID,
		TokenType: "v2",
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = profileName
	}
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Save credentials (private key path)
	creds, err := config.LoadCredentials()
	if err != nil {
		return err
	}
	creds.Set(profileName, config.Credentials{PrivateKeyFile: privateKeyFile})
	if err := creds.Save(); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	// Save token
	tokens, err := config.LoadTokens()
	if err != nil {
		return err
	}
	tokens.Set(profileName, config.TokenEntry{
		Token:     token,
		ExpiresAt: expiresAt,
		TokenType: "v2",
	})
	if err := tokens.Save(); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	jst := time.FixedZone("JST", 9*60*60)
	fmt.Fprintf(os.Stderr, "Logged in to profile %q with v2 token (expires %s / %s JST)\n",
		profileName,
		expiresAt.Format(time.RFC3339),
		expiresAt.In(jst).Format("2006-01-02 15:04"))
	verifyToken(token)
	return nil
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove token and credentials for the active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := getProfileFlag(cmd)

		tokens, err := config.LoadTokens()
		if err != nil {
			return err
		}
		tokens.Delete(profileName)
		if err := tokens.Save(); err != nil {
			return err
		}

		creds, err := config.LoadCredentials()
		if err != nil {
			return err
		}
		creds.Delete(profileName)
		if err := creds.Save(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Logged out of profile %q\n", profileName)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		profileName := getProfileFlag(cmd)
		profile, ok := cfg.Profiles[profileName]
		if !ok {
			fmt.Fprintf(os.Stderr, "Profile %q: not configured\n", profileName)
			return &lmerrors.ConfigError{Message: fmt.Sprintf("profile %q not found", profileName)}
		}

		tokens, err := config.LoadTokens()
		if err != nil {
			return err
		}

		fmt.Printf("Profile:    %s\n", profileName)
		fmt.Printf("Channel ID: %s\n", profile.ChannelID)
		fmt.Printf("Token Type: %s\n", profile.TokenType)

		if entry, ok := tokens.Get(profileName); ok && entry.Token != "" {
			if profile.TokenType == "longterm" || entry.ExpiresAt.IsZero() {
				fmt.Printf("Token:      set (longterm, no expiry)\n")
			} else {
				jst := time.FixedZone("JST", 9*60*60)
				remaining := time.Until(entry.ExpiresAt)
				if remaining > 0 {
					fmt.Printf("Token:      valid (expires in %s, %s JST)\n",
						remaining.Truncate(time.Minute),
						entry.ExpiresAt.In(jst).Format("2006-01-02 15:04"))
				} else {
					fmt.Printf("Token:      expired (%s ago, was %s JST)\n",
						(-remaining).Truncate(time.Minute),
						entry.ExpiresAt.In(jst).Format("2006-01-02 15:04"))
				}
			}
		} else {
			fmt.Printf("Token:      none\n")
		}

		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		tokens, err := config.LoadTokens()
		if err != nil {
			return err
		}

		if len(cfg.Profiles) == 0 {
			fmt.Fprintln(os.Stderr, "No profiles configured. Run 'lm auth login' to create one.")
			return nil
		}

		for name, profile := range cfg.Profiles {
			marker := " "
			if name == cfg.ActiveProfile {
				marker = "*"
			}
			tokenStatus := "no token"
			if tokens.IsValid(name) {
				tokenStatus = "authenticated"
			} else if _, ok := tokens.Get(name); ok {
				entry, _ := tokens.Get(name)
				if entry.Token != "" {
					tokenStatus = "expired"
				}
			}
			fmt.Printf("%s %s\t%s\t%s\t%s\n", marker, name, profile.ChannelID, profile.TokenType, tokenStatus)
		}
		return nil
	},
}

var switchCmd = &cobra.Command{
	Use:   "switch <profile>",
	Short: "Switch active profile",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if _, ok := cfg.Profiles[name]; !ok {
			return &lmerrors.ConfigError{Message: fmt.Sprintf("profile %q not found", name)}
		}

		cfg.ActiveProfile = name
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Switched to profile %q\n", name)
		return nil
	},
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Print current token to stdout (for scripting)",
	RunE: func(cmd *cobra.Command, args []string) error {
		check, _ := cmd.Flags().GetBool("check")
		profileName := getProfileFlag(cmd)

		if check {
			tokens, err := config.LoadTokens()
			if err != nil {
				return &lmerrors.AuthError{Message: "cannot load tokens"}
			}
			if tokens.IsValid(profileName) {
				return nil
			}
			return &lmerrors.AuthError{Message: "token is not valid"}
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		creds, err := config.LoadCredentials()
		if err != nil {
			return err
		}

		tokens, err := config.LoadTokens()
		if err != nil {
			return err
		}

		token, err := api.EnsureToken(profileName, cfg, creds, tokens)
		if err != nil {
			return err
		}

		fmt.Print(token)
		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <profile>",
	Short: "Completely remove a profile",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		delete(cfg.Profiles, name)
		if cfg.ActiveProfile == name {
			cfg.ActiveProfile = ""
			for k := range cfg.Profiles {
				cfg.ActiveProfile = k
				break
			}
		}
		if err := cfg.Save(); err != nil {
			return err
		}

		creds, err := config.LoadCredentials()
		if err != nil {
			return err
		}
		creds.Delete(name)
		_ = creds.Save()

		tokens, err := config.LoadTokens()
		if err != nil {
			return err
		}
		tokens.Delete(name)
		_ = tokens.Save()

		fmt.Fprintf(os.Stderr, "Removed profile %q\n", name)
		return nil
	},
}

func verifyToken(token string) {
	fmt.Fprintln(os.Stderr, "Verifying token...")
	client, err := api.NewClient(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		return
	}
	botAPI := &api.BotAPI{Client: client}
	info, err := botAPI.GetInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: token saved but verification failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'lm bot info' to test your token manually.")
		return
	}
	fmt.Fprintf(os.Stderr, "Verified as @%s (%s)\n", info.BasicID, info.DisplayName)
}

func getProfileFlag(cmd *cobra.Command) string {
	if p, _ := cmd.Flags().GetString("profile"); p != "" {
		return p
	}
	if p := config.EnvOr(config.EnvProfile, ""); p != "" {
		return p
	}
	cfg, _ := config.Load()
	if cfg != nil && cfg.ActiveProfile != "" {
		return cfg.ActiveProfile
	}
	return "default"
}
