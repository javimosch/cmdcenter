package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const Version = "1.0.0"

const appName = "cmdcenter"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		handleStart()
	case "stop":
		handleStop()
	case "status":
		handleStatus()
	case "version":
		handleVersion()
	case "add":
		handleAddCommand()
	case "edit":
		handleEditCommand()
	case "remove":
		handleRemoveCommand()
	case "list":
		handleListCommands()
	case "run":
		handleRunCommand()
	case "init":
		handleInitConfig()
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func handleStart() {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	port := startCmd.Int("port", 3031, "Port for HTTP server")
	daemon := startCmd.Bool("daemon", false, "Run as daemon")
	startCmd.Parse(os.Args[2:])

	if *daemon {
		startDaemon(*port)
	} else {
		startServer(*port)
	}
}

func handleStop() {
	stopDaemon()
}

func handleStatus() {
	checkDaemonStatus()
}

func handleVersion() {
	fmt.Printf("cmdcenter v%s\n", Version)
}

func handleInitConfig() {
	configPath, err := getConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file already exists at %s\n", configPath)
		fmt.Print("Do you want to overwrite it? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Aborted")
			return
		}
	}

	config := Config{
		Title:    "Command Center",
		Subtitle: "Generic command execution dashboard",
		Commands: []Command{},
	}

	if err := saveConfigFile(&config); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config file created at %s\n", configPath)
	fmt.Println("Edit the file to customize your commands")
}

func handleAddCommand() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	id := addCmd.String("id", "", "Command ID (required)")
	name := addCmd.String("name", "", "Command name (required)")
	description := addCmd.String("description", "", "Command description")
	icon := addCmd.String("icon", "🎯", "Command icon (emoji)")
	command := addCmd.String("command", "", "Shell command to execute (required)")
	supportsArgs := addCmd.Bool("supports-args", false, "Enable argument support")
	addCmd.Parse(os.Args[2:])

	if *id == "" || *name == "" || *command == "" {
		fmt.Fprintf(os.Stderr, "Error: id, name, and command are required\n")
		fmt.Println("Usage: cmdcenter add --id <id> --name <name> --command <command> [--description <desc>] [--icon <icon>] [--supports-args]")
		os.Exit(1)
	}

	config, err := loadConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Println("Run 'cmdcenter init' to create a config file first")
		os.Exit(1)
	}

	// Check if ID already exists
	for _, cmd := range config.Commands {
		if cmd.ID == *id {
			fmt.Fprintf(os.Stderr, "Error: Command with ID '%s' already exists\n", *id)
			fmt.Println("Use 'cmdcenter edit' to modify existing commands")
			os.Exit(1)
		}
	}

	// Add new command
	newCommand := Command{
		ID:           *id,
		Name:         *name,
		Description:  *description,
		Icon:         *icon,
		Command:      *command,
		SupportsArgs: *supportsArgs,
	}

	config.Commands = append(config.Commands, newCommand)

	if err := saveConfigFile(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Added command: %s (%s)\n", *name, *id)
}

func handleEditCommand() {
	editCmd := flag.NewFlagSet("edit", flag.ExitOnError)
	id := editCmd.String("id", "", "Command ID to edit (required)")
	name := editCmd.String("name", "", "New command name")
	description := editCmd.String("description", "", "New command description")
	icon := editCmd.String("icon", "", "New command icon")
	command := editCmd.String("command", "", "New shell command")
	supportsArgs := editCmd.Bool("supports-args", false, "Enable argument support")
	editCmd.Parse(os.Args[2:])

	if *id == "" {
		fmt.Fprintf(os.Stderr, "Error: id is required\n")
		fmt.Println("Usage: cmdcenter edit --id <id> [--name <name>] [--command <command>] [--description <desc>] [--icon <icon>] [--supports-args]")
		os.Exit(1)
	}

	config, err := loadConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Find and update command
	found := false
	for i, cmd := range config.Commands {
		if cmd.ID == *id {
			found = true
			if *name != "" {
				config.Commands[i].Name = *name
			}
			if *description != "" {
				config.Commands[i].Description = *description
			}
			if *icon != "" {
				config.Commands[i].Icon = *icon
			}
			if *command != "" {
				config.Commands[i].Command = *command
			}
			config.Commands[i].SupportsArgs = *supportsArgs
			break
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "Error: Command with ID '%s' not found\n", *id)
		fmt.Println("Use 'cmdcenter list' to see available commands")
		os.Exit(1)
	}

	if err := saveConfigFile(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Updated command: %s\n", *id)
}

func handleRunCommand() {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	id := runCmd.String("id", "", "Command ID to execute (required)")
	args := runCmd.String("args", "", "Extra arguments to append to command")
	runCmd.Parse(os.Args[2:])

	if *id == "" {
		fmt.Fprintf(os.Stderr, "Error: id is required\n")
		fmt.Println("Usage: cmdcenter run --id <id> [--args <args>]")
		os.Exit(1)
	}

	config, err := loadConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Println("Run 'cmdcenter init' to create a config file first")
		os.Exit(1)
	}

	// Find command
	var cmdToRun *Command
	for i, cmd := range config.Commands {
		if cmd.ID == *id {
			cmdToRun = &config.Commands[i]
			break
		}
	}

	if cmdToRun == nil {
		fmt.Fprintf(os.Stderr, "Error: Command with ID '%s' not found\n", *id)
		fmt.Println("Use 'cmdcenter list' to see available commands")
		os.Exit(1)
	}

	// Build command
	fullCommand := cmdToRun.Command
	if *args != "" {
		fullCommand = fullCommand + " " + *args
	}

	fmt.Printf("Executing: %s\n", fullCommand)

	// Execute command
	execCmd := exec.Command("bash", "-c", fullCommand)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func handleRemoveCommand() {
	removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
	id := removeCmd.String("id", "", "Command ID to remove (required)")
	removeCmd.Parse(os.Args[2:])

	if *id == "" {
		fmt.Fprintf(os.Stderr, "Error: id is required\n")
		fmt.Println("Usage: cmdcenter remove --id <id>")
		os.Exit(1)
	}

	config, err := loadConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Find and remove command
	found := false
	var updatedCommands []Command
	for _, cmd := range config.Commands {
		if cmd.ID == *id {
			found = true
		} else {
			updatedCommands = append(updatedCommands, cmd)
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "Error: Command with ID '%s' not found\n", *id)
		fmt.Println("Use 'cmdcenter list' to see available commands")
		os.Exit(1)
	}

	config.Commands = updatedCommands

	if err := saveConfigFile(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Removed command: %s\n", *id)
}

func handleListCommands() {
	config, err := loadConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Println("Run 'cmdcenter init' to create a config file first")
		os.Exit(1)
	}

	fmt.Printf("Title: %s\n", config.Title)
	fmt.Printf("Subtitle: %s\n", config.Subtitle)
	fmt.Println("\nCommands:")
	
	if len(config.Commands) == 0 {
		fmt.Println("  No commands configured")
		return
	}

	for i, cmd := range config.Commands {
		fmt.Printf("  %d. %s (%s)\n", i+1, cmd.Name, cmd.ID)
		fmt.Printf("     Description: %s\n", cmd.Description)
		fmt.Printf("     Icon: %s\n", cmd.Icon)
		fmt.Printf("     Command: %s\n", cmd.Command)
		fmt.Println()
	}
}

func printHelp() {
	fmt.Println("cmdcenter - Generic Command Center Dashboard")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cmdcenter <command> [options]")
	fmt.Println()
	fmt.Println("Server Commands:")
	fmt.Println("  start       Start HTTP server (UI)")
	fmt.Println("  stop        Stop daemon server")
	fmt.Println("  status      Check daemon status")
	fmt.Println("  version     Show version information")
	fmt.Println()
	fmt.Println("Config Commands:")
	fmt.Println("  init        Initialize config file with defaults")
	fmt.Println("  add         Add a new command")
	fmt.Println("  edit        Edit an existing command")
	fmt.Println("  remove      Remove a command")
	fmt.Println("  list        List all configured commands")
	fmt.Println("  run         Execute a command by ID")
	fmt.Println()
	fmt.Println("Start Options:")
	fmt.Println("  -port int   Port for HTTP server (default 3031)")
	fmt.Println("  -daemon     Run as daemon (background)")
	fmt.Println()
	fmt.Println("Add Options:")
	fmt.Println("  --id string          Command ID (required)")
	fmt.Println("  --name string        Command name (required)")
	fmt.Println("  --command string     Shell command (required)")
	fmt.Println("  --description string Command description")
	fmt.Println("  --icon string        Command icon (emoji)")
	fmt.Println("  --supports-args      Enable argument support for this command")
	fmt.Println()
	fmt.Println("Edit Options:")
	fmt.Println("  --id string          Command ID to edit (required)")
	fmt.Println("  --name string        New command name")
	fmt.Println("  --command string     New shell command")
	fmt.Println("  --description string New command description")
	fmt.Println("  --icon string        New command icon")
	fmt.Println("  --supports-args      Enable argument support for this command")
	fmt.Println()
	fmt.Println("Remove Options:")
	fmt.Println("  --id string          Command ID to remove (required)")
	fmt.Println()
	fmt.Println("Run Options:")
	fmt.Println("  --id string          Command ID to execute (required)")
	fmt.Println("  --args string        Extra arguments to append to command")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cmdcenter start")
	fmt.Println("  cmdcenter start -port 3000")
	fmt.Println("  cmdcenter start -daemon")
	fmt.Println("  cmdcenter init")
	fmt.Println("  cmdcenter add --id disk-usage --name 'Disk Usage' --command 'df -h' --icon '💾'")
	fmt.Println("  cmdcenter add --id df --name 'Disk Free' --command 'df' --icon '💾' --supports-args")
	fmt.Println("  cmdcenter edit --id disk-usage --command 'df -h | grep /'")
	fmt.Println("  cmdcenter remove --id disk-usage")
	fmt.Println("  cmdcenter list")
	fmt.Println("  cmdcenter run --id df --args '-h'")
	fmt.Println("  cmdcenter stop")
	fmt.Println("  cmdcenter status")
	fmt.Println("  cmdcenter version")
}