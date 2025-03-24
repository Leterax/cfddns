# **cfddns** - Cloudflare Dynamic DNS Updater

**cfddns** is a lightweight Go-based web service that updates DNS records on Cloudflare dynamically. It provides an API to update **A** and **AAAA** records for a given domain.

## **Features**
- Update **A** and **AAAA** DNS records via API calls.
- Uses Cloudflare's API to check and modify DNS records.
- Simple health check endpoint.

## **Installation**

1. Install Go (if not already installed):
   ```bash
   go version
   ```
   If Go is not installed, download it from [golang.org](https://golang.org/dl/).

2. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/cfddns.git
   cd cfddns
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

## **Configuration**
The service runs on port **8080** by default. You can change it by setting the `PORT` environment variable.

```bash
export PORT=8000
go run main.go
```

## **Usage**
### **Update DNS Record**
Send a GET request to update a DNS record:
```
GET /?token=<API_TOKEN>&email=<EMAIL>&zone=<ZONE_NAME>&record=<SUBDOMAIN>&ipv4=<IPv4_ADDRESS>&ipv6=<IPv6_ADDRESS>
```
- **token** (required) – Cloudflare API token
- **email** (required) – Cloudflare account email
- **zone** (required) – Domain name (e.g., example.com)
- **record** (optional) – Subdomain (e.g., www), defaults to root domain
- **ipv4** (optional) – New IPv4 address
- **ipv6** (optional) – New IPv6 address

Example:
```bash
curl "http://localhost:8080/?token=your_api_token&email=your_email@example.com&zone=example.com&record=www&ipv4=192.168.1.1"
```

### **Health Check**
Check if the service is running:
```bash
curl "http://localhost:8080/healthz"
```

## **Deployment**
To run the service in the background, you can use:
```bash
nohup go run main.go &
```

Or build and run it as a binary:
```bash
go build -o cfddns
./cfddns
```

## **Notes**
- Ensure your Cloudflare API token has permission to manage DNS records.
- Only updates the record if the IP address has changed.

---

Enjoy seamless DNS updates with cfddns! 🚀
