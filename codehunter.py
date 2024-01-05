import argparse
import re
import requests
from tqdm import tqdm

def count_matches(urls, expressions):
    results = []
    for u in tqdm(urls, desc="Processing URLs"):
        count = 0
        try:
            response = requests.get(u)
            if response.status_code == 200:
                html_content = response.text
                for expr in expressions:
                    pattern = re.compile(expr)
                    count += len(pattern.findall(html_content))
        except requests.RequestException as e:
            print(f"Failed to access URL {u}. Error: {e}")
            count = -1

        results.append((u, count))

    return results

def main():
    parser = argparse.ArgumentParser(description="Regular expression match counter in URLs")
    parser.add_argument('-f', '--file', help="File with URLs", required=True)
    parser.add_argument('-r', '--regex', help="File with regular expressions", required=True)
    parser.add_argument('-o', '--output', help="Output file", required=True)
    args = parser.parse_args()

    try:
        with open(args.file, 'r') as f:
            urls = f.read().splitlines()
    except FileNotFoundError:
        print(f"File {args.file} not found.")
        return

    try:
        with open(args.regex, 'r') as f:
            expressions = f.read().splitlines()
    except FileNotFoundError:
        print(f"File {args.regex} not found.")
        return

    results = count_matches(urls, expressions)

    with open(args.output, 'w') as f:
        for url, count in results:
            f.write(f"{url} [{count} times]\n")

if __name__ == "__main__":
    main()
