import argparse
import requests
from urllib.parse import urlparse, urljoin
from bs4 import BeautifulSoup
import sys
import time

# Script Info
SCRIPT_NAME = "urlextractor.py"
SCRIPT_VERSION = "0.1 Beta"
SCRIPT_CREATOR = "Albert C"

def main():
    parser = argparse.ArgumentParser(
        description='Extract URLs from a given website.',
        epilog=f'Example usage: \npython3 {SCRIPT_NAME} -u https://domain.com -d -v -o output_file.txt'
    )
    parser.add_argument('-u', '--url', type=str, help='URL to extract links from', required=True)
    parser.add_argument('-v', '--verbose', action='store_true', help='Verbose mode')
    parser.add_argument('-o', '--output', type=str, help='Output file name')
    parser.add_argument('-d', '--deep', action='store_true', help='Perform deep crawl')

    args = parser.parse_args()

    base_url = args.url
    if not base_url.startswith(('http://', 'https://')):
        base_url = 'http://' + base_url

    try:
        response = requests.get(base_url)
        response.raise_for_status()
    except requests.RequestException as e:
        print(f"Failed to fetch URL: {e}")
        return

    soup = BeautifulSoup(response.content, 'html.parser')

    if args.output:
        output_file_name = args.output
    else:
        domain_name = urlparse(base_url).hostname
        output_file_name = domain_name.replace('www.', '') if domain_name else 'output.txt'

    visited_urls = set()
    urls_to_extract = set()
    sitemap_url = check_sitemap(base_url)
    
    if sitemap_url:
        urls_from_sitemap = extract_urls_from_sitemap(sitemap_url)
        for url in urls_from_sitemap:
            urls_to_extract.add(url)

    if not args.verbose:
        print("Starting extraction... ", end='', flush=True)
        animate_loading()

    extract_urls(base_url, visited_urls, args, output_file_name, urls_to_extract)

def animate_loading():
    animation = "|/-\\"
    for i in range(20):
        time.sleep(0.1)
        sys.stdout.write("\b" + animation[i % len(animation)])
        sys.stdout.flush()

def check_sitemap(url):
    try:
        response = requests.get(url)
        response.raise_for_status()
        soup = BeautifulSoup(response.content, 'html.parser')
        sitemap_tag = soup.find('link', {'rel': 'sitemap'})

        if sitemap_tag:
            sitemap_url = urljoin(url, sitemap_tag.get('href'))
            return sitemap_url
        else:
            return None
    except requests.RequestException:
        return None

def extract_urls_from_sitemap(sitemap_url):
    try:
        response = requests.get(sitemap_url)
        response.raise_for_status()
        soup = BeautifulSoup(response.content, 'xml')  # Assuming the sitemap is in XML format

        urls = []
        for loc in soup.find_all("loc"):
            urls.append(loc.text)

        return urls
    except requests.RequestException:
        return []

def extract_urls(url, visited_urls, args, output_file_name, urls_to_extract):
    if url in visited_urls:
        return
    visited_urls.add(url)

    try:
        response = requests.get(url)
        response.raise_for_status()
    except requests.RequestException:
        return

    soup = BeautifulSoup(response.content, 'html.parser')

    with open(output_file_name, 'a') as output_file:
        for link in soup.find_all('a', href=True):
            href = link.get('href')
            if '#' in href:
                continue

            parsed_url = urlparse(href)
            if parsed_url.hostname is None:
                continue

            full_url = urljoin(url, href)

            if parsed_url.hostname == urlparse(url).hostname or parsed_url.hostname.endswith('.' + urlparse(url).hostname):
                if args.verbose:
                    print(full_url)

                if full_url not in visited_urls and full_url not in urls_to_extract:
                    output_file.write(full_url + '\n')
                    visited_urls.add(full_url)
                    
                    if args.deep:  
                        extract_urls(full_url, visited_urls, args, output_file_name, urls_to_extract)

if __name__ == "__main__":
    print(f"Script Name: {SCRIPT_NAME} | Version: {SCRIPT_VERSION} | Creator: {SCRIPT_CREATOR}")
    main()
