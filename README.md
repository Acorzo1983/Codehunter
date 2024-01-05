# Usage Video Demo
https://youtu.be/dkBlcVMkgB8

# URL Regex Match Counter

URL Regex Match Counter is a Python script designed to count occurrences of various regular expressions within the content of provided URLs.

## Usage

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

### How to Use

Run the `codehunter.py` script with the following arguments:

```bash
python codehunter.py -f <file_with_URLs> -r <file_with_regex> -o <output_file>
```
-f/--file: File containing URLs to scan.
-r/--regex: File containing regular expressions to match.
-o/--output: Output file to store the results.

```bash
python codehunter.py -f urls.txt -r regex.txt -o results.txt
```

# Features

Simultaneously scans multiple URLs for various regex patterns.
Provides the count of matches per URL for each regex.

# File Structure

codehunter.py: Main Python script.

README.md: Instructions and information about the script.
requirements.txt: Contains necessary Python packages.

# Contribution

Contributions, issues, and feature requests are welcome! Feel free to check the issues page if you want to contribute.
