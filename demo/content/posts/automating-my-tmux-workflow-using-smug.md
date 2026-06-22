---
title: Automating my Tmux workflow using smug
description: I recently automated more of my development workflow — this time for tmux. I used a CLI tool called smug and wrote configurations for automatically setting up my tmux sessions.
tags:
  - productivity
  - golang
  - workflow
status: published
createdAt: 2024-10-14
publishedAt: 2024-10-14
updatedAt: 2024-10-14
---

I've been using [tmux](https://github.com/tmux/tmux/wiki) everyday for a couple of years now. I use a simple bash script, called \
tmux-sessionizer, with tmux that will create tmux sessions for my projects. The credit for this script goes to [ThePrimeagen](https://www.youtube.com/c/theprimeagen). The bash script needs [fzf](https://junegunn.github.io/fzf/) — a command line fuzzy finder. When I press `ctrl-f` in my terminal, `fzf` will show a list of my projects in directories that I've configured in the tmux-sessionizer script. When I select one, tmux will create a session for that project and switch to that directory.

Here's the script:

```bash
#!/usr/bin/env bash

if [[ $# -eq 1 ]]; then
    selected=$1
else
    selected=$(find ~/Experiments ~/Projects ~/Learning ~/OSS ~/Work -mindepth 1 -maxdepth 2 -type d | fzf)
fi

if [[ -z $selected ]]; then
    exit 0
fi

selected_name=$(basename "$selected" | tr . _)
tmux_running=$(pgrep tmux)

if [[ -z $TMUX ]] && [[ -z $tmux_running ]]; then
    tmux new-session -s $selected_name -c $selected
    exit 0
fi

if ! tmux has-session -t=$selected_name 2>/dev/null; then
    tmux new-session -ds $selected_name -c $selected
fi

tmux switch-client -t $selected_name
```

Then what I usually do is press `ctrl-a + c` to create multiple tmux windows. I usually create 3 windows, each with its own purpose — general purpose commands, git or package manager commands, and running the local server. I do this so many times a day. It quickly becomes tedious but I never really addressed it for a long time. I was aware of [tmuxinator](https://github.com/tmuxinator/tmuxinator) — a program written in Ruby that let's us manage tmux sessions using a configuration file. More recently, I heard about [smug](https://github.com/ivaaaan/smug), also a tmux session manager written in Go. I've been learning Go and I figured I'll give smug a shot. It'll be fun to look at the source code and see how a CLI tool is written in Go. So I wrote some YAML configurations for few of my projects and ran the commands. I don't think I'm going back to my old way.

## What does smug do?

Smug let's us create configuration files for each of our projects and when we run

```bash
smug start <project_name>
```

it'll create a session for our project, create all the windows and panes in that session, and here's the cool part, run the commands we specify in each of those windows and panes. For example, here's the configuration for one of my projects:

```yaml
session: herald
root: ~/Projects/useherald/herald/
before_start:
  - docker start herald
stop:
  - docker stop $(docker ps -q)
windows:
  - name: code
    layout: main-vertical
    commands:
      - nvim
  - name: git
    layout: main-vertical
    commands:
      - gitty
  - name: server
    layout: main-vertical
    commands:
      - bunny
      - moon run server:tidy
      - dunstify "Herald running at dashboard - localhost:3000 and server - localhost:4000"
      - moon run :dev
```

I'm instructing smug to create 3 named windows, specifying layouts for them, and what commands they should run.

The first window runs Neovim. The second window runs a CLI tool called [gitty](https://github.com/muesli/gitty). It uses the GitHub API to fetch the list of issues and pull requests from the repository. gitty is also written in Go.

The third window is for my development server. Here's what each command does:

- `bunny` is an alias for `bun i` which installs all the packages from `package.json`
- `moon run server:tidy` runs `go mod tidy` inside a `server` directory which contains my Go REST API. [Moon](https://moonrepo.dev/moon) is a task runner and a monorepo management tool.
- `dunstify` is a command in `dunst`\
  — a notification daemon for Linux-based distros. This command sends a \
  system notification telling me that my development server is up and \
  running.
- `moon run :dev` runs the development server for both the client and server directories in my monorepo.

The configuration also has `before_start` and `stop` settings. Smug will run the commands in `before_start` before creating the tmux sessions. I use it to start a Docker container that runs a PostgreSQL database. The commands under `stop` are for any clean up I want to do after stopping the tmux session. Here, I've set it up to stop my PostgreSQL container.

## Integrating smug into the sessionizer script

I can now modify my sessionizer script to run smug for projects with a smug configuration and if not, fallback to only creating a tmux session.

```bash
#!/usr/bin/env bash

# Get the list of smug projects
smug_projects=$(smug list)

if [[ $# -eq 1 ]]; then
    selected=$1
else
    selected=$(find ~/Experiments ~/Projects ~/Learning ~/OSS ~/Work -mindepth 1 -maxdepth 2 -type d | fzf)
fi

if [[ -z $selected ]]; then
    exit 0
fi

selected_name=$(basename "$selected" | tr . _)

# Check if the selected project is in the smug list
if echo "$smug_projects" | grep -q "^$selected_name$"; then
    # If it's a smug project, start it with smug
    smug start "$selected_name"
else
    # If it's not a smug project, use the original tmux logic
    tmux_running=$(pgrep tmux)

    if [[ -z $TMUX ]] && [[ -z $tmux_running ]]; then
        tmux new-session -s $selected_name -c $selected
        exit 0
    fi

    if ! tmux has-session -t=$selected_name 2>/dev/null; then
        tmux new-session -ds $selected_name -c $selected
    fi

    tmux switch-client -t $selected_name
fi
```

Now if I press `ctrl-f`, `fzf` will show me a list of my projects and when I select one with a smug configuration, it'll automatically set everything up for me. I can go into Neovim and start coding right away. How cool is that? I wish I'd done this sooner.
