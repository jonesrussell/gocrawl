# MultiSource Module

## Overview

The `multisource` module is designed to manage and crawl multiple web sources for content. It reads configuration from a YAML file (`sources.yml`), which specifies the sources to be crawled, including their URLs and the corresponding Elasticsearch indices where the content will be stored.

## Structure

The module consists of the following key components:

- **SourceConfig**: A struct that represents the configuration for a single source, including:
  - `Name`: The name of the source.
  - `URL`: The URL of the source to be crawled.
  - `Index`: The Elasticsearch index where the content will be stored.

- **MultiSource**: A struct that manages multiple sources. It includes:
  - `Sources`: A slice of `SourceConfig` that holds the configurations for all sources.

- **Functions**:
  - `NewMultiSource()`: Initializes a new `MultiSource` instance by reading and unmarshalling the `sources.yml` file.
  - `Start(ctx context.Context)`: Begins the crawling process for all configured sources.
  - `crawlSource(ctx context.Context, source SourceConfig)`: Handles the crawling logic for a single source, including making HTTP requests and processing responses.
  - `Stop()`: Halts the crawling process if necessary.

## Configuration

The `sources.yml` file should be structured as follows:

```yaml
sources:
  - name: "Source Name"
    url: "https://example.com"
    index: "example_index"
```

### Example `sources.yml`

```yaml
sources:
  - name: "Elliot Lake Today"
    url: "https://www.elliotlaketoday.com/opp-beat"
    index: "elliot_lake_articles"
  
  - name: "Espanola News"
    url: "https://www.myespanolanow.com/news/"
    index: "espanola_articles"

  - name: "Mid North Monitor"
    url: "https://www.midnorthmonitor.com/category/news/local-news/"
    index: "mid_north_articles"
```

## Usage

1. **Initialization**: The `MultiSource` module is initialized by calling `NewMultiSource()`, which reads the `sources.yml` file.
2. **Start Crawling**: Call the `Start(ctx context.Context)` method to begin crawling all configured sources.
3. **Crawling Logic**: The module will make HTTP requests to each source URL, check the response status, and print a success message. You can implement additional logic to parse the response body and index the content into Elasticsearch.

## Future Enhancements

### 1. Implement HTML Parsing
- **Objective**: Extract relevant content (e.g., articles, titles, dates) from the HTML response of each crawled page.
- **Tasks**:
  - Research and select an HTML parsing library (e.g., `goquery`, `colly`).
  - Implement a function to parse the HTML response and extract the desired elements.
  - Define a structure to hold the extracted content (e.g., title, body, publication date).
- **Considerations**:
  - Different sources may have different HTML structures; consider creating source-specific parsing functions if necessary.
  - Ensure that the parsing logic is robust against changes in the HTML structure of the sources.

### 2. Add Error Handling and Logging
- **Objective**: Improve the robustness of the crawling process by handling errors gracefully and logging important events.
- **Tasks**:
  - Implement error handling for network issues, parsing errors, and unexpected response formats.
  - Use a logging library (e.g., `zap`, `logrus`) to log errors, warnings, and informational messages.
  - Create a logging strategy that includes log levels (e.g., debug, info, error) and log formatting.
- **Considerations**:
  - Ensure that logs are written to a file or external logging service for better monitoring.
  - Consider implementing retries for transient errors (e.g., network timeouts).

### 3. Implement Indexing Logic
- **Objective**: Send the extracted content to Elasticsearch for indexing.
- **Tasks**:
  - Research the Elasticsearch Go client (`github.com/elastic/go-elasticsearch/v8`).
  - Implement a function to format the extracted content into the appropriate structure for Elasticsearch.
  - Create an indexing function that sends the formatted content to the specified index in Elasticsearch.
- **Considerations**:
  - Handle potential errors during indexing (e.g., connection issues, invalid data).
  - Consider implementing bulk indexing for efficiency if multiple articles are extracted from a single source.

### 4. Add Configuration for Parsing Rules
- **Objective**: Allow users to define parsing rules for different sources in the `sources.yml` file.
- **Tasks**:
  - Extend the `SourceConfig` struct to include parsing rules (e.g., CSS selectors for titles, bodies).
  - Implement logic to apply these rules during the parsing process.
- **Considerations**:
  - Provide documentation on how to define parsing rules in the configuration file.
  - Ensure backward compatibility with existing configurations.

### 5. Implement Rate Limiting and Throttling
- **Objective**: Control the rate of requests sent to each source to avoid overwhelming them.
- **Tasks**:
  - Implement a rate-limiting mechanism (e.g., using a token bucket algorithm).
  - Allow users to specify a rate limit in the `sources.yml` file for each source.
- **Considerations**:
  - Ensure that the rate-limiting logic is flexible and can be adjusted based on user needs.
  - Monitor the performance and adjust the rate limits as necessary.

### 6. Create Unit and Integration Tests
- **Objective**: Ensure the reliability and correctness of the `multisource` module through testing.
- **Tasks**:
  - Write unit tests for individual functions (e.g., parsing, indexing).
  - Create integration tests that simulate the entire crawling process with mock data.
- **Considerations**:
  - Use a testing framework (e.g., `testing`, `testify`) to facilitate writing and running tests.
  - Consider using a mock Elasticsearch instance for integration tests.

### 7. Documentation and Examples
- **Objective**: Provide clear documentation and examples for users to understand how to use the `multisource` module effectively.
- **Tasks**:
  - Update the `README.md` file with detailed usage instructions and examples.
  - Create example configurations and sample output to demonstrate the module's capabilities.
- **Considerations**:
  - Ensure that the documentation is kept up to date with any changes made to the module.
  - Consider creating a dedicated documentation site if the project grows larger.
