# kmd

Turn your shell history into a Makefile.

## Install

Download the binary:
```bash
curl -L https://github.com/abishek77s/kmd/raw/main/kmd -o kmd
chmod +x kmd
sudo mv kmd /usr/local/bin/
```

Or build from source:
```bash
git clone https://github.com/abishek77s/kmd.git
cd kmd/cli
go build
```

## Supported Shells

✓ Bash  
✓ Zsh  
✓ Fish

## Usage

```bash
kmd              # Show last 5 commands
kmd -n 7        # Show last 7 commands
```


## What it does

1. Reads your shell history (bash/zsh/fish)
2. Lets you select commands in order
3. Generates a Makefile with numbered steps

```makefile
all: step1 step2 step3 step4

step1:
	@echo "🔧 Step 1: docker build -t myapp:latest ."
	docker build -t myapp:latest .

step2:
	@echo "🔧 Step 2: aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com"
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

step3:
	@echo "🔧 Step 3: docker tag myapp:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/myapp:latest"
	docker tag myapp:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/myapp:latest

step4:
	@echo "🔧 Step 4: docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/myapp:latest"
	docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/myapp:latest
```

Run with `make all` or `make step1`, `make step2`, etc.
