# Moobee - Movie Ticket Booking System

Moobee is a full-featured movie ticket booking application built with Go, offering users an intuitive interface to browse movies, select seats, and make bookings.

## 🎬 Features

- **User Authentication System**: Secure login and registration functionality
- **Landing Page**: Attractive introduction to the platform for first-time visitors
- **Movie Browsing**: Clean grid layout to browse all available movies
- **Seat Selection**: Interactive seat map for choosing seats
- **Booking Management**: View, manage, and cancel bookings
- **Admin Dashboard**: Comprehensive management tools for administrators
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Search Functionality**: Find movies easily

## 🛠️ Technologies

- **Backend**: Go (Golang)
- **Database**: SQLite
- **Frontend**: HTML, CSS, JavaScript
- **Template Engine**: Go's html/template package
- **Authentication**: Session-based with secure password hashing

## 📋 Prerequisites

- Go 1.21 or higher
- SQLite

## 🚀 Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/moobee.git
   cd moobee
2. Install dependencies:
    ```bash
    go mod tidy
    ```

3. Run the application:
    ```bash
    go run main.go
    ```

4. Open your browser and navigate to:
    ```
    http://localhost:8080
    ```

## 📂 Project Structure

- `main.go`: Entry point of the application
- `models/`: Contains database models
- `controllers/`: Handles application logic
- `views/`: HTML templates for the frontend
- `static/`: Static files like CSS, JavaScript, and images
- `routes/`: Defines application routes

## 🤝 Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## 🙏 Acknowledgements

Movie poster images are used for educational purposes only
Special thanks to everyone who contributed to the development and testing