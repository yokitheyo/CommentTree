# 🌲 CommentTree - Hierarchical Comment System

## ✨ Features

- **Hierarchical Comment Structure**: Support for nested comments with the ability to reply at any level of nesting
- **Interactive Web Interface**: Beautiful and intuitive interface with dark theme support
- **Comment Search**: Fast search through comment content using PostgreSQL full-text search
- **Sorting**: Ability to sort comments by creation date (ascending/descending)
- **Pagination**: Support for paginated comment display
- **Branch Collapsing**: Ability to collapse/expand comment branches
- **Comment Management**: Ability to add, reply, and soft-delete comments
- **Asynchronous Loading**: Smooth operation without page reloads
- **Responsive Design**: Full mobile and tablet support
- **Structured Logging**: Detailed logging with Zerolog

## 🐳 Quick Start with Docker

### Prerequisites
- Docker & Docker Compose installed
- Port 8080 available (or modify docker-compose.yml)

### Running the Application

```bash
# Clone the repository
git clone https://github.com/yokitheyo/CommentTree.git
cd CommentTree

# Start all services (PostgreSQL, API)
docker-compose up --build

# Application will be available at http://localhost:8080
```

### Services in Docker Compose

| Service | Port | Details |
|---------|------|---------|
| **API** | 8080 | Main application server |
| **PostgreSQL** | 5432 | Database (commenttree) |

### Environment Variables

Database connection settings are defined in the `config.yaml` file:

```yaml
database:
  dsn: "postgres://postgres:postgres@db:5432/commenttree?sslmode=disable"
```

### Stopping the Application

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

## 📡 API Endpoints

### 1. **Web Interface**

```
GET /
```
Returns the main web interface (HTML).

### 2. **Getting Comments**

```
GET /comments?parent={id}&limit={limit}&offset={offset}&sort={asc/desc}
```

**Example:**
```
GET /comments?limit=10&offset=0&sort=desc
```

**Parameters:**
- `parent` (optional): Parent comment ID to get only children
- `limit` (optional): Number of comments per page (default 10)
- `offset` (optional): Offset for pagination (default 0)
- `sort` (optional): Sort direction (asc/desc, default asc)

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "parent_id": null,
    "content": "This is the first comment",
    "author": "John Doe",
    "created_at": "2026-01-28T12:00:00Z",
    "updated_at": null,
    "deleted": false,
    "children": [
      {
        "id": 2,
        "parent_id": 1,
        "content": "This is a reply to the first comment",
        "author": "Jane Smith",
        "created_at": "2026-01-28T12:15:00Z",
        "updated_at": null,
        "deleted": false,
        "children": []
      }
    ]
  }
]
```

---

### 3. **Creating a Comment**

```
POST /comments
Content-Type: application/json

{
  "parent_id": 1,        // Optional: Parent comment ID
  "author": "Author Name", // Required: Comment author's name
  "content": "Comment text" // Required: Comment content
}
```

**Response (201 Created):**
```json
{
  "id": 3,
  "parent_id": 1,
  "content": "Comment text",
  "author": "Author Name",
  "created_at": "2026-01-28T12:30:00Z",
  "updated_at": null,
  "deleted": false,
  "children": []
}
```

---

### 4. **Deleting a Comment**

```
DELETE /comments/{id}
```

**Response (204 No Content)** on successful deletion.

---

### 5. **Searching Comments**

```
GET /comments/search?query={query}&limit={limit}&offset={offset}
```

**Example:**
```
GET /comments/search?query=important&limit=5&offset=0
```

**Response (200 OK):**
```json
[
  {
    "id": 5,
    "parent_id": 2,
    "content": "This is a very important comment",
    "author": "Alex Ivanov",
    "created_at": "2026-01-28T13:00:00Z",
    "updated_at": null,
    "deleted": false,
    "children": []
  }
]
```
