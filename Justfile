# Binary name and path
binary_dir := env_var("HOME") + "/code/bilte"
binary_name := "bilte"
cmd_path := "cmd"

# Build flags for smaller binary size
ldflags := "-w -s"
build_flags := "-trimpath -buildvcs=false -ldflags '" + ldflags + "'"

build:
    @echo "Building binary..."
    go build {{build_flags}} -o {{binary_dir}}/{{binary_name}} ./{{cmd_path}}

clean:
    @echo "Cleaning up..."
    rm -f {{binary_dir}}/{{binary_name}}

test:
    @echo "Running tests..."
    go test -v ./...

modernize:
    @echo "Running gopls modernize..."
    go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix -test ./...

lint:
    @echo "Running golangci-lint..."
    golangci-lint run ./...

sec:
    @echo "Running gosec..."
    gosec ./...

vuln:
    @echo "Running govulncheck..."
    govulncheck ./...

templ:
    @echo "Building templates..."
    go tool templ generate
