# go-via Communication Diagram

This diagram shows the communication flow between components in the go-via VMware imaging appliance.

## System Overview

```mermaid
flowchart TB
    subgraph "Frontend Layer"
        User[👤 User]
        Angular[🅰️ Angular SPA]
    end
    
    subgraph "Backend Layer"
        Router[🔧 Gin Router]
        API[📡 API Handlers]
    end
    
    subgraph "Data Layer"
        DB[(🗄️ SQLite Database)]
        Files[📂 File System]
    end
    
    subgraph "External Services"
        ILO[🖥️ ILO/Redfish API]
        TFTP[📁 TFTP Server]
    end
    
    subgraph "Real-time"
        WS[🔗 WebSocket Server]
    end
    
    User --> Angular
    Angular <--> Router
    Router --> API
    API <--> DB
    API <--> Files
    API <--> ILO
    API --> WS
    WS --> Angular
    TFTP <--> Files
    TFTP <--> API
```

## Main Communication Flows

### 1. User Authentication Flow
```mermaid
sequenceDiagram
    participant U as User
    participant A as Angular
    participant R as Router
    participant API as API
    participant DB as Database
    
    U->>A: Login Request
    A->>R: POST /v1/login
    R->>API: Login Handler
    API->>DB: Validate Credentials
    DB-->>API: User Data
    API-->>R: JWT Token
    R-->>A: Auth Response
    A-->>U: Login Success
```

### 2. Resource Management Flow
```mermaid
sequenceDiagram
    participant A as Angular
    participant R as Router
    participant API as API
    participant DB as Database
    
    Note over A,DB: CRUD Operations (Pools, Hosts, Groups, Images)
    
    A->>R: GET /v1/[resource]
    R->>API: List Handler
    API->>DB: Query Data
    DB-->>API: Results
    API-->>R: JSON Response
    R-->>A: Resource List
    
    A->>R: POST /v1/[resource]
    R->>API: Create Handler
    API->>DB: Insert Record
    DB-->>API: Created Item
    API-->>R: JSON Response
    R-->>A: Success Response
```

### 3. Host Management & ILO Operations
```mermaid
sequenceDiagram
    participant A as Angular
    participant R as Router
    participant API as API
    participant ILO as ILO/Redfish
    
    Note over A,ILO: Hardware Management Operations
    
    A->>R: POST /v1/ilohosts/checkilo
    R->>API: CheckIP Handler
    API->>ILO: Test Connection
    ILO-->>API: Status
    API-->>R: Result
    R-->>A: Connection Status
    
    A->>R: POST /v1/ilohosts/:id/start
    R->>API: Start Handler
    API->>ILO: Start Server
    ILO-->>API: Success/Failure
    API-->>R: Operation Result
    R-->>A: Start Response
```

### 4. Boot Process via TFTP
```mermaid
sequenceDiagram
    participant H as Host (PXE)
    participant T as TFTP Server
    participant API as API
    participant DB as Database
    participant F as File System
    
    H->>T: PXE Boot Request
    T->>F: Read Boot Files
    F-->>T: Boot Files
    T->>API: Generate Kickstart
    API->>DB: Get Host Config
    DB-->>API: Configuration
    API-->>T: Generated ks.cfg
    T-->>H: Boot Files + Config
```

### 5. Real-time Logging
```mermaid
sequenceDiagram
    participant A as Angular
    participant WS as WebSocket
    participant API as API
    
    A->>WS: Connect /v1/log
    WS-->>A: Connection Established
    
    loop Real-time Events
        API->>WS: Log Event
        WS->>A: Push Log Message
        A->>A: Display Log
    end
```

## API Endpoints Overview

### Core Resources
- **Authentication**: `/v1/login`
- **Pools**: `/v1/pools` (GET, POST, PATCH, DELETE)
- **Hosts**: `/v1/hosts` (GET, POST, PATCH, DELETE)
- **Groups**: `/v1/groups` (GET, POST, PATCH, DELETE)
- **Images**: `/v1/images` (GET, POST, DELETE)
- **Users**: `/v1/users` (GET, POST, PATCH, DELETE)

### Hardware Management
- **ILO Check**: `/v1/ilohosts/checkilo`
- **Host Control**: `/v1/ilohosts/:id/{start,shutdown,reboot}`
- **Boot Config**: `/v1/ilohosts/:id/{onetimeboot,setvlanID}`
- **Host Config**: `/v1/hostconfig`

### System Services
- **Theme**: `/v1/theme/image`
- **Logs**: `/v1/log` (WebSocket)
- **Version**: `/v1/version`
- **Kickstart**: `/ks.cfg`

## Component Responsibilities

### Frontend (Angular)
- **User Interface**: Forms, wizards, management screens
- **State Management**: Application state and routing
- **API Communication**: HTTP client for REST operations
- **Real-time Updates**: WebSocket connection for logs

### Backend (Go)
- **HTTP Server**: Gin router with middleware
- **Business Logic**: API handlers for all operations
- **Data Persistence**: SQLite database with GORM
- **File Management**: Static files and uploads
- **Hardware Integration**: ILO/Redfish API communication
- **Network Services**: TFTP server for boot files

### External Integrations
- **ILO/Redfish APIs**: Server hardware management
- **File System**: Boot images and configuration files
- **SQLite Database**: Application data storage
- **WebSocket**: Real-time event streaming
