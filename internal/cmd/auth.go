package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/HFunction2013/coj-cli/internal/api"
	"github.com/HFunction2013/coj-cli/internal/cli"
	"github.com/HFunction2013/coj-cli/internal/config"
	"github.com/HFunction2013/coj-cli/internal/term"
)

func authCmd() *cli.Command {
	c := &cli.Command{
		Name:        "auth",
		Aliases:     []string{"login"},
		Summary:     "Manage the login session",
		Description: "Log in, log out and inspect the current identity.\nThe session is stored locally as a PHPSESSID cookie.",
		Group:       "CORE COMMANDS",
	}
	c.Add(authLoginCmd(), authPhoneCmd(), authLogoutCmd(), authStatusCmd(), authRefreshCmd())
	return c
}

func authLoginCmd() *cli.Command {
	return &cli.Command{
		Name:    "login",
		Summary: "Log in with a username and password",
		Description: `Log in via GET /User/login and save the session.

Heads up: the endpoint is a GET and the password travels in the query
string (that is how the backend is built). Prefer a trusted network, or
use SMS login instead.`,
		Example: `  coj auth login --username alice --password secret
  coj auth login --username alice            # prompts for the password
  COJ_PASSWORD=secret coj auth login --username alice`,
		Flags: cli.NewFlagSet(
			&cli.Flag{Name: "username", Short: "u", Usage: "Account name"},
			&cli.Flag{Name: "password", Short: "p", Usage: "Password (prompted if omitted)"},
			&cli.Flag{Name: "skip-bind-phone", Usage: "Skip the bind-phone prompt", Bool: true},
		),
		Run: func(ctx *cli.Context) error {
			user := ctx.Flags.Get("username")
			if user == "" {
				user = os.Getenv("COJ_USERNAME")
			}
			if user == "" {
				return fmt.Errorf("missing --username (or set COJ_USERNAME)")
			}
			pwd := ctx.Flags.Get("password")
			if pwd == "" {
				pwd = os.Getenv("COJ_PASSWORD")
			}
			if pwd == "" {
				p, err := term.ReadPassword("Password: ")
				if err != nil {
					return err
				}
				pwd = p
			}

			c, err := clientFrom(ctx)
			if err != nil {
				return err
			}
			p := api.NewParams()
			p.Fields["user_name"] = user
			p.Fields["user_pwd"] = pwd
			if ctx.Flags.Bool("skip-bind-phone") {
				p.Fields["skip_bind_phone"] = "1"
			}

			ep := api.ByPath("/User/login")
			resp, err := c.Call(ep, p)
			if err != nil {
				return err
			}
			if err := resp.Check(); err != nil {
				return err
			}
			return persistLogin(ctx, c, resp, user)
		},
	}
}

func authPhoneCmd() *cli.Command {
	return &cli.Command{
		Name:    "sms",
		Summary: "Log in with a phone number and SMS code",
		Description: `Two steps: request a code with --send, then log in with --code.

    /User/sendCodeForLogin   request the code
    /User/loginByPhone       verify and log in`,
		Example: `  coj auth sms --phone 13800138000 --send
  coj auth sms --phone 13800138000 --code 123456`,
		Flags: cli.NewFlagSet(
			&cli.Flag{Name: "phone", Usage: "Phone number"},
			&cli.Flag{Name: "send", Usage: "Only request the code", Bool: true},
			&cli.Flag{Name: "code", Usage: "SMS code"},
		),
		Run: func(ctx *cli.Context) error {
			phone := ctx.Flags.Get("phone")
			if phone == "" {
				return fmt.Errorf("missing --phone")
			}
			c, err := clientFrom(ctx)
			if err != nil {
				return err
			}
			if ctx.Flags.Bool("send") {
				p := api.NewParams()
				p.Fields["F_Phone"] = phone
				resp, err := c.Call(api.ByPath("/User/sendCodeForLogin"), p)
				if err != nil {
					return err
				}
				if err := resp.Check(); err != nil {
					return err
				}
				fmt.Fprintln(ctx.Stdout, "Verification code sent")
				return nil
			}
			code := ctx.Flags.Get("code")
			if code == "" {
				return fmt.Errorf("missing --code (or send one first with --send)")
			}
			p := api.NewParams()
			p.Fields["F_Phone"] = phone
			p.Fields["F_Code"] = code
			resp, err := c.Call(api.ByPath("/User/loginByPhone"), p)
			if err != nil {
				return err
			}
			if err := resp.Check(); err != nil {
				return err
			}
			return persistLogin(ctx, c, resp, phone)
		},
	}
}

func authLogoutCmd() *cli.Command {
	return &cli.Command{
		Name:    "logout",
		Summary: "Clear the stored session",
		Run: func(ctx *cli.Context) error {
			if err := config.ClearSession(); err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, "Logged out")
			return nil
		},
	}
}

func authStatusCmd() *cli.Command {
	return &cli.Command{
		Name:    "status",
		Summary: "Show the current login identity",
		Example: "  coj auth status\n",
		Run: func(ctx *cli.Context) error {
			c, err := clientFrom(ctx)
			if err != nil {
				return err
			}
			if c.SessionID() == "" {
				return fmt.Errorf("not logged in")
			}
			cfg, _ := config.Load()
			rows := []map[string]interface{}{
				{"key": "host", "value": c.BaseURL},
				{"key": "session", "value": c.SessionID()},
				{"key": "username", "value": cfg.Username},
				{"key": "role", "value": config.RoleName(cfg.UserType)},
			}
			// confirm the session is still valid server-side
			resp, err := c.Call(api.ByPath("/User/heartbeat"), api.NewParams())
			if err == nil && resp.OK() {
				rows = append(rows, map[string]interface{}{"key": "server", "value": "session valid"})
			} else {
				rows = append(rows, map[string]interface{}{"key": "server", "value": "session expired or unreachable"})
			}
			return emit(ctx, rows)
		},
	}
}

func authRefreshCmd() *cli.Command {
	return &cli.Command{
		Name:    "ping",
		Summary: "Send a heartbeat (refresh the server-side idle window)",
		Description: `Call GET /User/heartbeat.

The backend checkLogin() uses a 120-minute sliding window. By design the
heartbeat refreshes it, while polling endpoints such as searchInvite and
getCanMoveClassList only verify without extending it.`,
		Run: func(ctx *cli.Context) error {
			c, err := clientFrom(ctx)
			if err != nil {
				return err
			}
			if err := requireSession(ctx, c); err != nil {
				return err
			}
			resp, err := c.Call(api.ByPath("/User/heartbeat"), api.NewParams())
			if err != nil {
				return err
			}
			if err := resp.Check(); err != nil {
				return err
			}
			fmt.Fprintln(ctx.Stdout, "Heartbeat sent")
			return nil
		},
	}
}

// persistLogin stores the session and prints the identity.
func persistLogin(ctx *cli.Context, c *api.Client, resp *api.Response, account string) error {
	if err := config.SaveSession(c.SessionID()); err != nil {
		return err
	}
	cfg, _ := config.Load()
	cfg.Host = c.BaseURL
	cfg.Username = account

	var info map[string]interface{}
	if err := json.Unmarshal(resp.Data, &info); err == nil {
		if v, ok := info["F_Type"]; ok {
			switch t := v.(type) {
			case float64:
				cfg.UserType = int(t)
			case string:
				fmt.Sscanf(t, "%d", &cfg.UserType)
			}
		}
		if v, ok := info["F_UserName"].(string); ok && v != "" {
			cfg.Username = v
		}
	}
	if err := cfg.Save(); err != nil {
		stderrf(ctx, "warning: could not save config: %v", err)
	}

	fmt.Fprintf(ctx.Stdout, "Logged in as %s (%s)\n", cfg.Username, config.RoleName(cfg.UserType))
	return nil
}
