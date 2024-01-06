# Usage Video Demo
https://youtu.be/dkBlcVMkgB8

# URL Regex Match Counter

URL Regex Match Counter is a script designed to count occurrences of various regular expressions within the content of provided URLs. It is available in both Python and Go versions.

## Usage - Python Version

### Prerequisites

- Python 3.x
- Required Python packages: `requests`, `tqdm`

### Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/Acorzo1983/codehunter.git
    cd codehunter
    ```

2. Install the required Python packages:
    ```bash
    pip install -r requirements.txt
    ```

### How to Use - Python Version

Run the `codehunter.py` script with the following arguments:

```bash
python codehunter.py -f <file_with_URLs> -r <file_with_regex> -v -o <output_file>
```

-f/--file: File containing URLs to scan.

-r/--regex: File containing regular expressions to match.

-o/--output: Output file to store the results.

-v/--verbose: Optional flag for verbose output (displays URL results).


```bash
python codehunter.py -f urls.txt -r regex.txt -o results.txt
```

### Usage - Go Version

Prerequisites
Go installed on your machine.

How to Use
1. Compile the Go code:

```bash
go build codehunter.go
```

2. Run the compiled executable with the necessary arguments:

```bash
./codehunter -f <file_with_URLs> -r <file_with_regex> -v -o <output_file>
```

### Features
Simultaneously scans multiple URLs for various regex patterns.
Provides the count of matches per URL for each regex.

### File Structure
codehunter.py: Main Python script.

codehunter.go: Go version of the script.

README.md: Instructions and information about the script.

requirements.txt: Contains necessary Python packages.

Contribution
Contributions, issues, and feature requests are welcome! Feel free to check the issues page if you want to contribute.

```bash
This README includes instructions for both the Python and Go versions of the script. Feel free to adjust it further or add more details as needed.
```


