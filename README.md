# CodeHunter v2.5

**Ultra-Fast Bug Bounty Scanner for Kali Linux**  
*Made with ❤️ by Albert.C ([@yz9yt](https://twitter.com/yz9yt))*

---

## What is CodeHunter?

CodeHunter is a lightning-fast pattern hunting tool designed for Bug Bounty hunters and penetration testers. It scans URLs for specific patterns (secrets, API endpoints, admin panels, etc.) and outputs only the URLs where matches are found.

Perfect integration with Kali Linux, Parrot OS, and popular Bug Bounty tools like `katana`, `subfinder`, `httpx`, and `waybackurls`.

---

## Why CodeHunter?

- ⚡ **Ultra-Fast** – Multi-threaded scanning with Go  
- 🧠 **Specialized Patterns** – 325+ professional regex patterns  
- 🔁 **Pipe-Friendly** – Perfect integration with Bug Bounty workflows  
- 🐱‍💻 **Kali Ready** – Built for penetration testing distributions  
- 🛠 **Simple** – One command, powerful results  

---

## Quick Install

### One-Line Installer

```bash
curl -sSL https://raw.githubusercontent.com/Acorzo1983/Codehunter/main/installer.sh | bash
```

### Manual Installation

```bash
# Clone repository
git clone https://github.com/Acorzo1983/Codehunter.git
cd Codehunter

# Build and install
make install

# Or just build locally
make build
```

---

## Usage

### Basic Usage

```bash
# Scan URLs for secrets
codehunter -r secrets.txt -l urls.txt -o found.txt

# Verbose output
codehunter -r api_endpoints.txt -l urls.txt -v

# Pipe from stdin
cat urls.txt | codehunter -r admin_panels.txt
```

### Real Bug Bounty Workflows

#### Finding Secrets in JavaScript

```bash
waybackurls tesla.com | grep "\.js$" | codehunter -r js_secrets.txt -o tesla_secrets.txt
```

#### Complete Subdomain to Secrets Workflow

```bash
subfinder -d tesla.com | httpx -mc 200,301,302 > live_subs.txt
cat live_subs.txt | katana -d 3 > all_urls.txt
codehunter -r secrets.txt -l all_urls.txt -o critical_findings.txt
codehunter -r api_endpoints.txt -l all_urls.txt -o api_endpoints.txt
codehunter -r admin_panels.txt -l all_urls.txt -o admin_panels.txt
```

#### Anonymous Scanning with Proxychains

```bash
proxychains codehunter -r secrets.txt -l targets.txt -o found.txt
```

#### Pipeline with Multiple Tools

```bash
subfinder -d tesla.com | httpx | katana | codehunter -r secrets.txt,api_endpoints.txt -v
```

#### Mobile App Testing

```bash
katana -u https://mobile-api.tesla.com | codehunter -r api_endpoints.txt -o mobile_apis.txt
```

---

## Pattern Files

| Pattern File      | Patterns | Description                      | Use Case                       |
|-------------------|----------|----------------------------------|--------------------------------|
| `secrets.txt`     | 72       | API keys, tokens, credentials    | Finding leaked secrets         |
| `api_endpoints.txt` | 45     | REST APIs, GraphQL, microservices | API discovery                 |
| `admin_panels.txt` | 58      | Admin areas, CMS panels, tools   | Finding admin access           |
| `js_secrets.txt`  | 65       | JavaScript secrets, configs      | Client-side secret hunting     |
| `files.txt`       | 85       | Sensitive files, backups, configs| File discovery                 |
| `custom.txt`      | Template | User customizable patterns       | Custom hunting                 |

---

## Pattern Examples

```bash
codehunter -r secrets.txt -l urls.txt
codehunter -r api_endpoints.txt -l urls.txt
codehunter -r admin_panels.txt -l urls.txt
codehunter -r js_secrets.txt -l js_urls.txt
codehunter -r files.txt -l urls.txt
codehunter -r secrets.txt,api_endpoints.txt,admin_panels.txt -l urls.txt
```

---

## Command Line Options

```text
-r string    Patterns file (required)
             Example: secrets.txt, api_endpoints.txt

-l string    URLs file (optional, uses stdin if not provided)
             Example: urls.txt, targets.txt

-o string    Output file (optional, uses stdout if not provided)
             Example: found.txt, results.txt

-t int       Number of threads (default 10)
             Example: -t 20 for faster scanning

-v           Verbose output (shows scanning progress)
-b           Show banner (default true)
```

---

## Build from Source

### Requirements

- Go 1.21+  
- Linux/macOS (Windows not supported)  
- Make (optional)  

### Build Commands

```bash
git clone https://github.com/Acorzo1983/Codehunter.git
cd Codehunter
make dev
make build
make install
make test
make clean
make help
```

---

## Supported Platforms

| Platform      | Status           | Notes           |
|---------------|------------------|------------------|
| Kali Linux    | Primary Target   | Fully optimized  |
| Parrot OS     | Fully Supported  | Native support   |
| Ubuntu/Debian | Supported        | Tested           |
| Arch Linux    | Supported        | AUR compatible   |
| macOS         | Compatible       | Intel/ARM64      |
| Windows       | Not Supported    | Use WSL2 instead |

---

## Real-World Examples

### Enterprise Bug Bounty

```bash
subfinder -d company.com | httpx | katana -d 2 | codehunter -r secrets.txt,api_endpoints.txt -o enterprise_findings.txt
```

### Mobile API Hunting

```bash
echo "https://api.mobile-app.com" | katana | codehunter -r api_endpoints.txt -v
```

### Cloud Infrastructure

```bash
waybackurls target.com | grep -E "(aws|gcp|azure)" | codehunter -r secrets.txt -o cloud_secrets.txt
```

### JavaScript Analysis

```bash
cat domains.txt | httpx | katana | grep "\.js$" | codehunter -r js_secrets.txt -o js_findings.txt
```

### Anonymous Recon

```bash
proxychains subfinder -d target.com | proxychains httpx | proxychains codehunter -r secrets.txt
```

---

## Performance

**Benchmarks (tested on Kali Linux)**

- Speed: 1000+ URLs/minute  
- Memory: <50MB RAM usage  
- Threads: Configurable (default: 10)  
- Patterns: 325+ regex patterns  
- Accuracy: Low false positive rate  

### Optimization Tips

```bash
codehunter -r secrets.txt -l urls.txt -t 20
codehunter -r api_endpoints.txt -l urls.txt
katana -u target.com | codehunter -r secrets.txt
```

---

## Legal & Responsible Usage

**Important Disclaimers**

- Only use on authorized targets  
- Respect bug bounty program rules  
- Follow responsible disclosure  
- Obtain proper authorization  
- Do not use for illegal activities  

**Intended Users**

- Bug bounty hunters  
- Penetration testers  
- Security researchers  
- Red team operators  
- Authorized security assessments  

---

## Advanced Configuration

### Custom Pattern Creation

```bash
nano patterns/custom.txt
echo "custom_api_key\s*[=:]\s*[a-zA-Z0-9]{32}" >> patterns/custom.txt
codehunter -r custom.txt -l urls.txt
```

### Integration Scripts

```bash
#!/bin/bash
# Bug bounty automation script

DOMAIN=$1
echo "[+] Starting recon for $DOMAIN"

# Subdomain discovery
subfinder -d $DOMAIN | httpx > live_hosts.txt

# URL discovery
cat live_hosts.txt | katana > all_urls.txt

# Hunt patterns
codehunter -r secrets.txt -l all_urls.txt -o secrets_found.txt
codehunter -r api_endpoints.txt -l all_urls.txt -o apis_found.txt
codehunter -r admin_panels.txt -l all_urls.txt -o admin_found.txt

echo "[+] Hunt complete! Check *_found.txt files"
```

---

## Contributing

**Contributions are welcome!**

### Code Contributions

- Fork the repository  
- Create a feature branch  
- Make your changes  
- Test thoroughly  
- Submit a pull request  

### Pattern Contributions

- Add patterns to appropriate files in `patterns/`  
- Test with real data  
- Ensure low false positive rate  
- Document the pattern purpose  

### Bug Reports

- Use GitHub Issues  
- Include OS and Go version  
- Provide reproduction steps  
- Include sample data (sanitized)  

---

## Support & Contact

- **GitHub:** [https://github.com/Acorzo1983/Codehunter](https://github.com/Acorzo1983/Codehunter)  
- **Twitter:** [@yz9yt](https://twitter.com/yz9yt)  
- **Issues:** [GitHub Issues](https://github.com/Acorzo1983/Codehunter/issues)  

---

## License

**MIT License** – See the LICENSE file for details.

**Summary:**  
- Commercial use ✅  
- Modification ✅  
- Distribution ✅  
- Private use ✅  
- No liability ❌  
- No warranty ❌  

---

## Credits & Acknowledgments

**Creator**: Made with ❤️ by Albert.C ([@yz9yt](https://twitter.com/yz9yt))

**Special Thanks**

- Kali Linux Team – For the amazing platform  
- Bug Bounty Community – For inspiration and feedback  
- Go Team – For the fantastic language  
- Open Source Contributors – For making this possible  

**Inspired By**

- Real-world bug bounty experiences  
- Penetration testing best practices  
- Community feedback and needs  
- Kali Linux tool ecosystem  

---

## What's Next?

**Planned Features**

- JSON output format  
- Custom HTTP headers support  
- Rate limiting options  
- Pattern hit statistics  
- Cloud storage integration  
- API key validation  

**Ideas & Requests**  
Have an idea? Open an issue or contribute!

---

**Happy Bug Hunting!**  
**CodeHunter v2.5 – Made with ❤️ by Albert.C ([@yz9yt](https://twitter.com/yz9yt))**  
**GitHub**: https://github.com/Acorzo1983/Codehunter  
**Star this repo if CodeHunter helps you find bugs!**
