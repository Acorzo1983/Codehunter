import argparse
import re
import requests
from tqdm import tqdm

# ANSI escape codes for text colors
GREEN = '\033[92m'
ORANGE = '\033[33m'
RED = '\033[91m'
END = '\033[0m'

def count_matches(urls, expressions, verbose, output_file):
    results = []

    for u in tqdm(urls, desc=f"{ORANGE}Code Hunter Running", bar_format="{l_bar}{bar}{r_bar}", disable=verbose):
        matches = {expr: 0 for expr in expressions}
        try:
            response = requests.get(u)
            if response.status_code == 200:
                html_content = response.text
                for expr in expressions:
                    pattern = re.compile(expr)
                    matches[expr] = len(pattern.findall(html_content))
        except requests.RequestException as e:
            print(f"Failed to access URL {u}. Error: {e}")

        results.append((u, matches))

        if verbose:
            display_result(u, matches)

    if output_file:
        with open(output_file, 'w') as f:
            for url, matches in results:
                f.write(f"{url} ")
                for expr, count in matches.items():
                    expr_formatted = expr.capitalize()
                    if count > 0:
                        f.write(f"-- [{GREEN}{expr_formatted} - {count} Times{END}] ")
                    else:
                        f.write(f"-- [{RED}{expr_formatted} - No Results{END}] ")
                f.write("\n")

def display_result(url, matches):
    print(f"{url} ", end="")
    for expr, count in matches.items():
        expr_formatted = expr.capitalize()
        if count > 0:
            print(f"-- [{GREEN}{expr_formatted} - {count} Times{END}] ", end="")
        else:
            print(f"-- [{RED}{expr_formatted} - No Results{END}] ", end="")
    print()

def main():
    print(f"""
{ORANGE} 
 ______            __         _______                __                
|      |.-----..--|  |.-----.|   |   |.--.--..-----.|  |_ .-----..----.
|   ---||  _  ||  _  ||  -__||       ||  |  ||     ||   _||  -__||   _|
|______||_____||_____||_____||___|___||_____||__|__||____||_____||__|  

{END}""")
    print("CodeHunter Version 1.1 by Albert C\n")  # Display version and credits
    parser = argparse.ArgumentParser(description="Regular expression match counter in URLs")
    parser.add_argument('-f', '--file', help="File with URLs", required=True)
    parser.add_argument('-r', '--regex', help="File with regular expressions", required=True)
    parser.add_argument('-o', '--output', help="Output file")
    parser.add_argument('-v', '--verbose', help="Verbose output", action='store_true')
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

    count_matches(urls, expressions, args.verbose, args.output)

if __name__ == "__main__":
    main()
