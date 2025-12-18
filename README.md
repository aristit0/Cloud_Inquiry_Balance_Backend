# Inquiry Balance API - Backend

Backend API menggunakan Golang untuk inquiry balance dengan Couchbase Cloud Capella.

## Prerequisites

- Go 1.21 atau lebih tinggi
- Python 3.8+ (untuk data generator)
- Akses ke Couchbase Cloud Capella

## Setup Data Generator (Python)

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Jalankan script untuk generate 10000 data:
```bash
python generate_data.py
```

Script ini akan membuat:
- 10000 account records di collection `ddmast`
- 10000 customer records di collection `cif`
- Setiap account ter-link ke CIF customer

## Setup Backend (Golang)

1. Install dependencies:
```bash
go mod download
```

2. Build aplikasi:
```bash
go build -o inquiry-balance-api
```

3. Jalankan server:
```bash
./inquiry-balance-api
```

Server akan berjalan di `http://localhost:8080`

## API Endpoints

### 1. Inquiry Balance
**POST** `/api/v1/inquiry`

Request Body:
```json
{
  "account": "000000001"
}
```

Response Success (200):
```json
{
  "response_code": "200",
  "response_message": "Success",
  "timestamp": "2024-12-18T10:30:00Z",
  "account": {
    "account_number": "000000001",
    "account_name": "John Doe",
    "cif": "CIF0000001",
    "account_type": "SAVINGS",
    "currency": "IDR",
    "available_balance": 15000000.00,
    "hold_balance": 0.00,
    "status": "ACTIVE",
    "branch_code": "BR123",
    "open_date": "2020-01-15",
    "last_transaction_date": "2024-12-17"
  },
  "customer": {
    "cif": "CIF0000001",
    "customer_type": "INDIVIDUAL",
    "full_name": "John Doe",
    "date_of_birth": "1990-05-15",
    "id_type": "KTP",
    "id_number": "3201234567890123",
    "email": "john.doe@email.com",
    "phone": "+62211234567",
    "mobile": "+628123456789",
    "address": {
      "street": "Jl. Sudirman No. 123",
      "city": "Jakarta",
      "province": "DKI Jakarta",
      "postal_code": "12345",
      "country": "Indonesia"
    },
    "customer_segment": "RETAIL",
    "kyc_status": "VERIFIED"
  }
}
```

Response Not Found (404):
```json
{
  "response_code": "404",
  "response_message": "Account not found",
  "timestamp": "2024-12-18T10:30:00Z"
}
```

Response Error (400):
```json
{
  "response_code": "400",
  "response_message": "Account number is required",
  "timestamp": "2024-12-18T10:30:00Z"
}
```

### 2. Health Check
**GET** `/api/v1/health`

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-12-18T10:30:00Z",
  "service": "inquiry-balance-api",
  "version": "1.0.0"
}
```

## Testing dengan curl

```bash
# Test inquiry balance
curl -X POST http://localhost:8080/api/v1/inquiry \
  -H "Content-Type: application/json" \
  -d '{"account": "000000001"}'

# Test health check
curl http://localhost:8080/api/v1/health
```

## Testing dengan Postman

1. Import collection atau buat request baru
2. Method: POST
3. URL: `http://localhost:8080/api/v1/inquiry`
4. Headers: `Content-Type: application/json`
5. Body (raw JSON):
```json
{
  "account": "000000001"
}
```

## Error Handling

API menangani berbagai error conditions:
- **400**: Bad Request (invalid format, missing account)
- **404**: Account Not Found
- **500**: Internal Server Error
- **200**: Success (dengan atau tanpa customer data)

## Notes

- Account numbers are zero-padded 9 digits (e.g., "000000001")
- CIF format: "CIF" + zero-padded 7 digits (e.g., "CIF0000001")
- Server menggunakan CORS untuk allow semua origins (untuk development)
- Connection menggunakan WAN development profile untuk optimasi latency