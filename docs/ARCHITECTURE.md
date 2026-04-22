Tech Stack

Frontend

* Next.js
* TypeScript
* Tailwind CSS
* SWR
* Recharts

Backend

* Go
* Gin
* PostgreSQL
* Redis

Machine Learning

* Python
* FastAPI
* scikit-learn
* pandas
* numpy

Infrastructure

* Docker
* Docker Compose
* Nginx

High-Level Architecture
flowchart LR

    A[Next.js Frontend] --> B[Go API]

    B --> C[PostgreSQL]

    B --> D[Redis]

    B --> E[Python ML Service]

    E --> B