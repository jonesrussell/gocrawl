# Go Web Crawler

A simple web crawler built in Go that fetches and processes web pages. This project demonstrates the use of the Colly library for web scraping and the Zap library for logging.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Tasks](#tasks)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- Crawl web pages with configurable depth and rate limits.
- Log crawling activities using structured logging.
- Easily configurable through environment variables.

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/jonesrussell/gocrawl.git
   cd gocrawl
   ```

2. **Install dependencies:**

   Make sure you have Go installed. Then, run:

   ```bash
   go mod tidy
   ```

3. **Create a `.env` file** (optional):

   You can create a `.env` file in the root directory to set environment variables for configuration.

## Usage

To run the web crawler, use the following command:

```bash
task crawl
```

You can customize the parameters such as `maxDepth`, `rateLimit`, and `url` directly in the command.

## Tasks

This project uses a `Taskfile.yml` for task automation. Here are the available tasks:

- **crawl**: Run the web crawler with specified parameters.
- **build**: Build the web crawler executable and place it in the `bin` directory.
- **lint**: Lint the Go code using `go vet` and `golangci-lint`.
- **test**: Run the tests for the application.

### Example of Building the Application

To build the application, run:

```bash
task build
```

This will create an executable named `gocrawl` in the `bin` directory.

## Testing

To run the tests, use:

```bash
task test
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
