package main

// Update the homeTemplate to support logged in users
const homeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Moobee - Movie Ticket Booking</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <header>
        <div class="logo">Moobee</div>
        <p>Your Ultimate Movie Experience</p>
    </header>
    <nav class="navbar">
        <div class="nav-left">
            <a href="/home" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
            {{if .User}}
                <a href="/bookings">My Bookings</a>
                <a href="/profile">Profile</a>
                {{if .User.IsAdmin}}
                    <a href="/admin">Admin</a>
                {{end}}
            {{end}}
        </div>
        <div class="nav-right">
            {{if .User}}
                <span class="welcome-text">Welcome, {{.User.Name}}</span>
                <a href="/logout" class="nav-btn logout-btn">Logout</a>
            {{else}}
                <a href="/login" class="nav-btn login-btn">Login</a>
                <a href="/register" class="nav-btn signup-btn">Sign Up</a>
            {{end}}
        </div>
    </nav>
    
    <main class="container">
        <div class="search-form">
            <form action="/search" method="get">
                <div class="form-group">
                    <input type="text" name="q" class="form-control" placeholder="Search for movies...">
                    <button type="submit" class="btn">Search</button>
                </div>
            </form>
        </div>
        
        <h2>Now Showing</h2>
        
        <div class="movie-grid">
            {{range .Movies}}
            <div class="movie-card">
                <img src="{{.Image}}" alt="{{.Title}}" class="movie-image">
                <div class="movie-details">
                    <h3 class="movie-title">{{.Title}}</h3>
                    <div class="movie-info">
                        <span><strong>Showtime:</strong> {{.Time}}</span>
                        <span><strong>Duration:</strong> {{.Duration}}</span>
                        <span><strong>Price:</strong> {{formatPrice .Price}}</span>
                        <span><strong>Available seats:</strong> {{availableSeats .Seats}}</span>
                    </div>
                    <a href="/book/{{.ID}}" class="btn">Book Now</a>
                </div>
            </div>
            {{end}}
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2023 Moobee. All rights reserved.</p>
        </div>
    </footer>
</body>
</html>`

// Add all the missing templates
const bookTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Book Tickets - {{.Movie.Title}} | Moobee</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <header>
        <div class="logo">Moobee</div>
        <p>Your Ultimate Movie Experience</p>
    </header>
    <nav class="navbar">
        <div class="nav-left">
            <a href="/home" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
            {{if and .User .User.ID}}
                <a href="/bookings">My Bookings</a>
                <a href="/profile">Profile</a>
                {{if .User.IsAdmin}}
                    <a href="/admin">Admin</a>
                {{end}}
            {{end}}
        </div>
        <div class="nav-right">
            {{if .User}}
                <span class="welcome-text">Welcome, {{.User.Name}}</span>
                <a href="/logout" class="nav-btn logout-btn">Logout</a>
            {{else}}
                <a href="/login" class="nav-btn login-btn">Login</a>
                <a href="/register" class="nav-btn signup-btn">Sign Up</a>
            {{end}}
        </div>
    </nav>
    
    <main class="container">
        <h2>Book Tickets for "{{.Movie.Title}}"</h2>
        
        <div class="booking-container">
            <div class="movie-preview">
                <img src="{{.Movie.Image}}" alt="{{.Movie.Title}}" class="movie-poster">
                <div class="movie-info-box">
                    <p><strong>Showtime:</strong> {{.Movie.Time}}</p>
                    <p><strong>Duration:</strong> {{.Movie.Duration}}</p>
                    <p><strong>Price:</strong> {{formatPrice .Movie.Price}} per seat</p>
                    <p><strong>Available seats:</strong> {{availableSeats .Movie.Seats}}</p>
                </div>
            </div>
            
            <div class="seat-selection">
                <h3>Select Your Seats</h3>
                
                <div class="legend">
                    <div class="legend-item">
                        <div class="seat"></div>
                        <span>Available</span>
                    </div>
                    <div class="legend-item">
                        <div class="seat selected"></div>
                        <span>Selected</span>
                    </div>
                    <div class="legend-item">
                        <div class="seat booked"></div>
                        <span>Booked</span>
                    </div>
                </div>
                
                <div class="screen">SCREEN</div>
                
                <div class="seat-grid">
                    {{range $rowIndex, $row := .Movie.Seats}}
                        {{range $colIndex, $isBooked := $row}}
                            <div class="seat{{if $isBooked}} booked{{end}}" data-row="{{$rowIndex}}" data-col="{{$colIndex}}">
                                {{$rowIndex}}-{{$colIndex}}
                            </div>
                        {{end}}
                    {{end}}
                </div>
            </div>
            
            <div class="booking-form-container">
                <h3>Booking Information</h3>
                
                <div id="selected-seats-list" class="selected-seats-summary"></div>
                
                <form id="booking-form" class="form">
                    <div class="form-group">
                        <label for="name">Full Name</label>
                        <input type="text" id="name" name="name" class="form-control" value="{{.User.Name}}" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="email">Email Address</label>
                        <input type="email" id="email" name="email" class="form-control" value="{{.User.Email}}" required>
                    </div>
                    
                    <input type="hidden" id="movieID" value="{{.Movie.ID}}">
                    <input type="hidden" id="price" value="{{.Movie.Price}}">
                    
                    <div class="form-group total-price">
                        <p><strong>Total: <span id="total">$0.00</span></strong></p>
                    </div>
                    
                    <button type="submit" class="btn" id="book-btn" disabled>Complete Booking</button>
                </form>
            </div>
        </div>
        
        <div id="booking-result"></div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2023 Moobee. All rights reserved.</p>
        </div>
    </footer>
    
    <script>
        // Keep the JavaScript functionality the same but fix the selector to use modern syntax
        document.addEventListener('DOMContentLoaded', function() {
            const selectedSeats = new Set();
            const price = parseFloat(document.getElementById('price').value);
            
            function updateTotal() {
                const total = selectedSeats.size * price;
                document.getElementById('total').textContent = '$' + total.toFixed(2);
                
                // Update the selected seats list
                const list = document.getElementById('selected-seats-list');
                if (selectedSeats.size > 0) {
                    let html = '<p><strong>Selected Seats:</strong> ';
                    html += Array.from(selectedSeats).join(', ');
                    html += '</p>';
                    list.innerHTML = html;
                    document.getElementById('book-btn').disabled = false;
                } else {
                    list.innerHTML = '<p>Please select at least one seat.</p>';
                    document.getElementById('book-btn').disabled = true;
                }
            }
            
            // Initialize seats
            document.querySelectorAll('.seat').forEach(seat => {
                if (!seat.classList.contains('booked')) {
                    seat.addEventListener('click', function() {
                        const row = this.getAttribute('data-row');
                        const col = this.getAttribute('data-col');
                        const seatId = row + '-' + col;
                        
                        if (this.classList.contains('selected')) {
                            this.classList.remove('selected');
                            selectedSeats.delete(seatId);
                        } else {
                            this.classList.add('selected');
                            selectedSeats.add(seatId);
                        }
                        
                        updateTotal();
                    });
                }
            });
            
            // Handle form submission
            document.getElementById('booking-form').addEventListener('submit', function(e) {
                e.preventDefault();
                
                if (selectedSeats.size === 0) {
                    alert('Please select at least one seat.');
                    return;
                }
                
                const name = document.getElementById('name').value.trim();
                const email = document.getElementById('email').value.trim();
                const movieID = document.getElementById('movieID').value;
                
                if (!name || !email) {
                    alert('Please provide your name and email.');
                    return;
                }
                
                // Disable the book button to prevent multiple submissions
                document.getElementById('book-btn').disabled = true;
                
                // Create booking request
                const bookingData = {
                    name: name,
                    email: email,
                    movieID: parseInt(movieID),
                    seats: Array.from(selectedSeats)
                };
                
                // Send booking request
                fetch('/api/book', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(bookingData)
                })
                .then(response => response.json())
                .then(data => {
                    if (data.Success) {
                        // Booking successful
                        document.getElementById('booking-result').innerHTML = 
                            '<div class="alert alert-success">' +
                            '<h3>Booking Successful!</h3>' +
                            '<p>' + data.Message + '</p>' +
                            '<p>Your booking ID: ' + data.BookingID + '</p>' +
                            '<p><a href="/booking/' + data.BookingID + '" class="btn">View Booking</a></p>' +
                            '</div>';
                            
                        // Mark the selected seats as booked
                        selectedSeats.forEach(seatId => {
                            const [row, col] = seatId.split('-');
                            const seat = document.querySelector('.seat[data-row="' + row + '"][data-col="' + col + '"]');
                            seat.classList.remove('selected');
                            seat.classList.add('booked');
                            seat.replaceWith(seat.cloneNode(true));
                        });
                        
                        // Clear the selection
                        selectedSeats.clear();
                        updateTotal();
                    } else {
                        // Booking failed
                        document.getElementById('booking-result').innerHTML = 
                            '<div class="alert alert-danger">' +
                            '<h3>Booking Failed</h3>' +
                            '<p>' + data.Message + '</p>' +
                            '</div>';
                            
                        // Re-enable the book button
                        document.getElementById('book-btn').disabled = false;
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    document.getElementById('booking-result').innerHTML = 
                        '<div class="alert alert-danger">' +
                        '<h3>Error</h3>' +
                        '<p>An error occurred while processing your booking. Please try again.</p>' +
                        '</div>';
                        
                    // Re-enable the book button
                    document.getElementById('book-btn').disabled = false;
                });
            });
        });
    </script>
</body>
</html>`

const bookingsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>My Bookings - CinemaGo</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <header>
        <h1>CinemaGo</h1>
    </header>
    <nav class="navbar">
        <div class="nav-left">
            <a href="/home" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
            {{if .User}}
                <a href="/bookings">My Bookings</a>
                <a href="/profile">Profile</a>
                {{if .User.IsAdmin}}
                    <a href="/admin">Admin</a>
                {{end}}
            {{end}}
        </div>
        <div class="nav-right">
            {{if .User}}
                <span class="welcome-text">Welcome, {{.User.Name}}</span>
                <a href="/logout" class="nav-btn logout-btn">Logout</a>
            {{else}}
                <a href="/login" class="nav-btn login-btn">Login</a>
                <a href="/register" class="nav-btn signup-btn">Sign Up</a>
            {{end}}
        </div>
    </nav>
    
    <main class="container">
        <h2>My Bookings</h2>
        
        {{if .Bookings}}
            <div class="bookings-list">
                {{range .Bookings}}
                    {{$movie := getMovie .MovieID}}
                    <div class="booking-item">
                        <div class="booking-header">
                            <h3>{{if $movie}}{{$movie.Title}}{{else}}Movie ID: {{.MovieID}}{{end}}</h3>
                            <span>Booked on {{.Date.Format "Jan 2, 2006 at 3:04 PM"}}</span>
                        </div>
                        <div class="booking-details">
                            <p><strong>Seats:</strong> {{range .Seats}}{{.}} {{end}}</p>
                            <p><strong>Name:</strong> {{.Name}}</p>
                            <p><strong>Email:</strong> {{.Email}}</p>
                            <p><strong>Total:</strong> {{formatPrice .Total}}</p>
                        </div>
                        <div class="booking-actions">
                            <a href="/booking/{{.ID}}" class="btn">View Details</a>
                            <a href="/cancel/{{.ID}}" class="btn btn-danger" onclick="return confirm('Are you sure you want to cancel this booking?')">Cancel Booking</a>
                        </div>
                    </div>
                {{end}}
            </div>
        {{else}}
            <p>You haven't made any bookings yet.</p>
            <p><a href="/" class="btn">Browse Movies</a></p>
        {{end}}
    </main>
</body>
</html>`

const viewBookingTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Booking Details - CinemaGo</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <header>
        <h1>CinemaGo</h1>
    </header>
    <nav class="navbar">
        <div class="nav-left">
            <a href="/home" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
            {{if .User}}
                <a href="/bookings">My Bookings</a>
                <a href="/profile">Profile</a>
                {{if .User.IsAdmin}}
                    <a href="/admin">Admin</a>
                {{end}}
            {{end}}
        </div>
        <div class="nav-right">
            {{if .User}}
                <span class="welcome-text">Welcome, {{.User.Name}}</span>
                <a href="/logout" class="nav-btn logout-btn">Logout</a>
            {{else}}
                <a href="/login" class="nav-btn login-btn">Login</a>
                <a href="/register" class="nav-btn signup-btn">Sign Up</a>
            {{end}}
        </div>
    </nav>
    
    <main class="container">
        <h2>Booking Details</h2>
        
        <div class="booking-details-card">
            <div class="booking-header">
                <h3>Booking #{{.Booking.ID}}</h3>
                <span>Booked on {{.Booking.Date.Format "Jan 2, 2006 at 3:04 PM"}}</span>
            </div>
            
            {{$movie := getMovie .Booking.MovieID}}
            <div class="movie-info">
                <img src="{{if $movie}}{{$movie.Image}}{{else}}/static/images/default.jpg{{end}}" alt="Movie Poster" class="movie-image-small">
                <div>
                    <h3>{{if $movie}}{{$movie.Title}}{{else}}Movie ID: {{.Booking.MovieID}}{{end}}</h3>
                    {{if $movie}}
                        <p><strong>Showtime:</strong> {{$movie.Time}}</p>
                        <p><strong>Duration:</strong> {{$movie.Duration}}</p>
                    {{end}}
                </div>
            </div>
            
            <div class="booking-details">
                <h4>Customer Information</h4>
                <p><strong>Name:</strong> {{.Booking.Name}}</p>
                <p><strong>Email:</strong> {{.Booking.Email}}</p>
                
                <h4>Seats</h4>
                <div class="seat-list">
                    {{range .Booking.Seats}}
                        <div class="seat-tag">{{.}}</div>
                    {{end}}
                </div>
                
                <h4>Payment</h4>
                <p><strong>Total:</strong> {{formatPrice .Booking.Total}}</p>
            </div>
            
            <div class="booking-actions">
                <a href="/bookings" class="btn">Back to My Bookings</a>
                <a href="/cancel/{{.Booking.ID}}" class="btn btn-danger" onclick="return confirm('Are you sure you want to cancel this booking?')">Cancel Booking</a>
            </div>
        </div>
    </main>
</body>
</html>`

const loginTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Login - Moobee</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        .auth-container {
            max-width: 450px;
            margin: 2rem auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
            overflow: hidden;
            animation: fadeIn 0.8s ease-out;
        }
        
        .auth-header {
            background: linear-gradient(135deg, var(--primary), var(--primary-light));
            color: white;
            padding: 2rem;
            text-align: center;
        }
        
        .auth-header h2 {
            margin-bottom: 0.5rem;
        }
        
        .auth-header p {
            opacity: 0.8;
        }
        
        .auth-body {
            padding: 2.5rem;
        }
        
        .form-group {
            margin-bottom: 1.5rem;
        }
        
        .form-group label {
            display: block;
            margin-bottom: 0.5rem;
            font-weight: 500;
            color: var(--dark);
        }
        
        .form-control {
            width: 100%;
            padding: 1rem 1.2rem;
            border: 1px solid #e1e1e1;
            border-radius: 8px;
            font-size: 1rem;
            transition: all 0.3s ease;
        }
        
        .form-control:focus {
            border-color: var(--primary);
            box-shadow: 0 0 0 3px rgba(255, 71, 87, 0.1);
        }
        
        .btn-auth {
            width: 100%;
            padding: 1rem;
            border: none;
            border-radius: 8px;
            background: var(--primary);
            color: white;
            font-size: 1rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            margin-top: 1rem;
        }
        
        .btn-auth:hover {
            background: var(--primary-light);
            transform: translateY(-3px);
            box-shadow: 0 5px 15px rgba(255, 71, 87, 0.3);
        }
        
        .auth-footer {
            text-align: center;
            margin-top: 1.5rem;
            color: #666;
        }
        
        .auth-footer a {
            color: var(--primary);
            text-decoration: none;
            font-weight: 500;
        }
        
        .auth-footer a:hover {
            text-decoration: underline;
        }
        
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
    </style>
</head>
<body>
    <header>
        <div class="logo">Moobee</div>
        <p>Your Ultimate Movie Experience</p>
    </header>
    
    <nav class="navbar">
        <div class="nav-left">
            <a href="/" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
        </div>
        <div class="nav-right">
            <a href="/register" class="nav-btn signup-btn">Sign Up</a>
        </div>
    </nav>
    
    <main class="container">
        <div class="auth-container">
            <div class="auth-header">
                <h2>Welcome Back</h2>
                <p>Log in to continue to Moobee</p>
            </div>
            
            <div class="auth-body">
                {{if .Error}}
                <div class="alert alert-danger">
                    {{.Error}}
                </div>
                {{end}}
                
                <form method="post">
                    <div class="form-group">
                        <label for="email">Email Address</label>
                        <input type="email" id="email" name="email" class="form-control" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="password">Password</label>
                        <input type="password" id="password" name="password" class="form-control" required>
                    </div>
                    
                    <button type="submit" class="btn-auth">Login</button>
                </form>
                
                <div class="auth-footer">
                    Don't have an account? <a href="/register">Create one now</a>
                </div>
            </div>
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2023 Moobee. All rights reserved.</p>
        </div>
    </footer>
</body>
</html>`

const registerTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Register - Moobee</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        .auth-container {
            max-width: 450px;
            margin: 2rem auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
            overflow: hidden;
            animation: fadeIn 0.8s ease-out;
        }
        
        .auth-header {
            background: linear-gradient(135deg, var(--primary), var(--primary-light));
            color: white;
            padding: 2rem;
            text-align: center;
        }
        
        .auth-header h2 {
            margin-bottom: 0.5rem;
        }
        
        .auth-header p {
            opacity: 0.8;
        }
        
        .auth-body {
            padding: 2.5rem;
        }
        
        .auth-footer {
            text-align: center;
            margin-top: 1.5rem;
            color: #666;
        }
        
        .auth-footer a {
            color: var(--primary);
            text-decoration: none;
            font-weight: 500;
        }
        
        .auth-footer a:hover {
            text-decoration: underline;
        }
        
        .btn-auth {
            width: 100%;
            padding: 1rem;
            border: none;
            border-radius: 8px;
            background: var(--primary);
            color: white;
            font-size: 1rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            margin-top: 1rem;
        }
        
        .btn-auth:hover {
            background: var(--primary-light);
            transform: translateY(-3px);
            box-shadow: 0 5px 15px rgba(255, 71, 87, 0.3);
        }
        
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
    </style>
</head>
<body>
    <header>
        <div class="logo">Moobee</div>
        <p>Your Ultimate Movie Experience</p>
    </header>
    
    <nav class="navbar">
        <div class="nav-left">
            <a href="/" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
        </div>
        <div class="nav-right">
            <a href="/login" class="nav-btn login-btn">Login</a>
        </div>
    </nav>
    
    <main class="container">
        <div class="auth-container">
            <div class="auth-header">
                <h2>Create an Account</h2>
                <p>Join Moobee for the ultimate movie experience</p>
            </div>
            
            <div class="auth-body">
                {{if .Error}}
                <div class="alert alert-danger">
                    {{.Error}}
                </div>
                {{end}}
                
                <form method="post">
                    <div class="form-group">
                        <label for="name">Full Name</label>
                        <input type="text" id="name" name="name" class="form-control" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="email">Email Address</label>
                        <input type="email" id="email" name="email" class="form-control" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="password">Password</label>
                        <input type="password" id="password" name="password" class="form-control" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="password2">Confirm Password</label>
                        <input type="password" id="password2" name="password2" class="form-control" required>
                    </div>
                    
                    <button type="submit" class="btn-auth">Create Account</button>
                </form>
                
                <div class="auth-footer">
                    Already have an account? <a href="/login">Login</a>
                </div>
            </div>
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2023 Moobee. All rights reserved.</p>
        </div>
    </footer>
</body>
</html>`

const profileTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>My Profile - CinemaGo</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <header>
        <h1>CinemaGo</h1>
    </header>
    <nav class="navbar">
        <div class="nav-left">
            <a href="/home" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
            {{if .User}}
                <a href="/bookings">My Bookings</a>
                <a href="/profile">Profile</a>
                {{if .User.IsAdmin}}
                    <a href="/admin">Admin</a>
                {{end}}
            {{end}}
        </div>
        <div class="nav-right">
            {{if .User}}
                <span class="welcome-text">Welcome, {{.User.Name}}</span>
                <a href="/logout" class="nav-btn logout-btn">Logout</a>
            {{else}}
                <a href="/login" class="nav-btn login-btn">Login</a>
                <a href="/register" class="nav-btn signup-btn">Sign Up</a>
            {{end}}
        </div>
    </nav>
    
    <main class="container">
        <h2>My Profile</h2>
        
        <div class="card">
            <div class="card-body">
                <h3>Account Information</h3>
                <p><strong>Name:</strong> {{.User.Name}}</p>
                <p><strong>Email:</strong> {{.User.Email}}</p>
                <p><strong>Member Since:</strong> {{.User.DateCreated.Format "January 2, 2006"}}</p>
                
                {{if .User.IsAdmin}}
                <p><span class="badge">Admin</span></p>
                {{end}}
            </div>
        </div>
        
        <h3>My Recent Bookings</h3>
        <div class="bookings-list">
            {{if .Bookings}}
                {{range .Bookings}}
                    {{$movie := getMovie .MovieID}}
                    <div class="booking-item">
                        <div class="booking-header">
                            <h4>{{if $movie}}{{$movie.Title}}{{else}}Movie ID: {{.MovieID}}{{end}}</h4>
                            <span>{{.Date.Format "Jan 2, 2006 at 3:04 PM"}}</span>
                        </div>
                        <div class="booking-details">
                            <p><strong>Seats:</strong> {{range .Seats}}{{.}} {{end}}</p>
                            <p><strong>Total:</strong> {{formatPrice .Total}}</p>
                        </div>
                        <div class="booking-actions">
                            <a href="/booking/{{.ID}}" class="btn">View Ticket</a>
                        </div>
                    </div>
                {{end}}
            {{else}}
                <p>You haven't made any bookings yet.</p>
                <p><a href="/" class="btn">Browse Movies</a></p>
            {{end}}
        </div>
    </main>
</body>
</html>`

// Update admin template header
const adminTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Admin Dashboard - Moobee</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
    <header>
        <div class="logo">Moobee Admin</div>
        <p>Management Dashboard</p>
    </header>
    <nav class="navbar">
        <div class="nav-left">
            <a href="/home" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
            {{if .User}}
                <a href="/bookings">My Bookings</a>
                <a href="/profile">Profile</a>
                {{if .User.IsAdmin}}
                    <a href="/admin">Admin</a>
                {{end}}
            {{end}}
        </div>
        <div class="nav-right">
            {{if .User}}
                <span class="welcome-text">Welcome, {{.User.Name}}</span>
                <a href="/logout" class="nav-btn logout-btn">Logout</a>
            {{else}}
                <a href="/login" class="nav-btn login-btn">Login</a>
                <a href="/register" class="nav-btn signup-btn">Sign Up</a>
            {{end}}
        </div>
    </nav>
    
    <main class="container">
        <h2>Admin Dashboard</h2>
        
        <div class="admin-stats">
            <div class="stat-card">
                <h3>Total Movies</h3>
                <p class="stat-value">{{.MovieCount}}</p>
            </div>
            
            <div class="stat-card">
                <h3>Total Bookings</h3>
                <p class="stat-value">{{.BookingCount}}</p>
            </div>
            
            <div class="stat-card">
                <h3>Total Revenue</h3>
                <p class="stat-value">{{formatPrice .TotalRevenue}}</p>
            </div>
            
            <div class="stat-card">
                <h3>Total Users</h3>
                <p class="stat-value">{{.UserCount}}</p>
            </div>
        </div>
        
        <h2>Recent Bookings</h2>
        <div class="bookings-list">
            {{range .RecentBookings}}
                {{$movie := getMovie .MovieID}}
                <div class="booking-item">
                    <div class="booking-header">
                        <h4>{{if $movie}}{{$movie.Title}}{{else}}Movie ID: {{.MovieID}}{{end}}</h4>
                        <span>{{.Date.Format "Jan 2, 2006 at 3:04 PM"}}</span>
                    </div>
                    <div class="booking-details">
                        <p><strong>Customer:</strong> {{.Name}} ({{.Email}})</p>
                        <p><strong>Seats:</strong> {{range .Seats}}{{.}} {{end}}</p>
                        <p><strong>Total:</strong> {{formatPrice .Total}}</p>
                    </div>
                </div>
            {{end}}
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2023 Moobee. All rights reserved.</p>
        </div>
    </footer>
</body>
</html>`

const adminMoviesTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Manage Movies - CinemaGo</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <header>
        <h1>CinemaGo Admin</h1>
    </header>
    <nav class="navbar">
        <div class="nav-left">
            <a href="/home" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
            {{if .User}}
                <a href="/bookings">My Bookings</a>
                <a href="/profile">Profile</a>
                {{if .User.IsAdmin}}
                    <a href="/admin">Admin</a>
                {{end}}
            {{end}}
        </div>
        <div class="nav-right">
            {{if .User}}
                <span class="welcome-text">Welcome, {{.User.Name}}</span>
                <a href="/logout" class="nav-btn logout-btn">Logout</a>
            {{else}}
                <a href="/login" class="nav-btn login-btn">Login</a>
                <a href="/register" class="nav-btn signup-btn">Sign Up</a>
            {{end}}
        </div>
    </nav>
    
    <main class="container">
        <h2>Manage Movies</h2>
        
        <div class="card">
            <div class="card-header">
                <h3>Add New Movie</h3>
            </div>
            <div class="card-body">
                <form method="post" class="form" enctype="multipart/form-data">
                    <input type="hidden" name="id" value="">
                    
                    <div class="form-group">
                        <label for="title">Movie Title</label>
                        <input type="text" id="title" name="title" class="form-control" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="time">Showtime</label>
                        <input type="text" id="time" name="time" class="form-control" placeholder="YYYY-MM-DD HH:MM" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="duration">Duration</label>
                        <input type="text" id="duration" name="duration" class="form-control" placeholder="2h 30m" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="image">Movie Poster</label>
                        <input type="file" id="image" name="image" class="form-control" accept="image/*">
                    </div>
                    
                    <div class="form-group">
                        <label for="price">Ticket Price</label>
                        <input type="number" id="price" name="price" step="0.01" class="form-control" required>
                    </div>
                    
                    <button type="submit" class="btn">Add Movie</button>
                </form>
            </div>
        </div>
        
        <h3>Current Movies</h3>
        <div class="movie-grid">
            {{range .}}
            <div class="movie-card">
                <img src="{{.Image}}" alt="{{.Title}}" class="movie-image">
                <div class="movie-details">
                    <h3 class="movie-title">{{.Title}}</h3>
                    <div class="movie-info">
                        <span><strong>Showtime:</strong> {{.Time}}</span>
                        <span><strong>Duration:</strong> {{.Duration}}</span>
                        <span><strong>Price:</strong> {{formatPrice .Price}}</span>
                        <span><strong>Available seats:</strong> {{availableSeats .Seats}}</span>
                    </div>
                    <div class="movie-actions">
                        <a href="/book/{{.ID}}" class="btn">View</a>
                        <a href="/admin/movies/delete/{{.ID}}" class="btn btn-danger" onclick="return confirm('Are you sure you want to delete this movie?')">Delete</a>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
    </main>
</body>
</html>`

const searchTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Search Results - CinemaGo</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <header>
        <h1>CinemaGo</h1>
    </header>
    <nav class="navbar">
        <div class="nav-left">
            <a href="/home" class="nav-logo">Moobee</a>
        </div>
        <div class="nav-links">
            <a href="/home">Movies</a>
            {{if .User}}
                <a href="/bookings">My Bookings</a>
                <a href="/profile">Profile</a>
                {{if .User.IsAdmin}}
                    <a href="/admin">Admin</a>
                {{end}}
            {{end}}
        </div>
        <div class="nav-right">
            {{if .User}}
                <span class="welcome-text">Welcome, {{.User.Name}}</span>
                <a href="/logout" class="nav-btn logout-btn">Logout</a>
            {{else}}
                <a href="/login" class="nav-btn login-btn">Login</a>
                <a href="/register" class="nav-btn signup-btn">Sign Up</a>
            {{end}}
        </div>
    </nav>
    
    <main class="container">
        <h2>Search Results for "{{.Query}}"</h2>
        
        <form action="/search" method="get" class="search-form">
            <div class="form-group">
                <input type="text" name="q" class="form-control" value="{{.Query}}" placeholder="Search for movies...">
                <button type="submit" class="btn">Search</button>
            </div>
        </form>
        
        <div class="movie-grid">
            {{range .Movies}}
            <div class="movie-card">
                <img src="{{.Image}}" alt="{{.Title}}" class="movie-image">
                <div class="movie-details">
                    <h3 class="movie-title">{{.Title}}</h3>
                    <div class="movie-info">
                        <span><strong>Showtime:</strong> {{.Time}}</span>
                        <span><strong>Duration:</strong> {{.Duration}}</span>
                        <span><strong>Price:</strong> {{formatPrice .Price}}</span>
                        <span><strong>Available seats:</strong> {{availableSeats .Seats}}</span>
                    </div>
                    <a href="/book/{{.ID}}" class="btn">Book Now</a>
                </div>
            </div>
            {{else}}
            <p>No movies found matching your search.</p>
            {{end}}
        </div>
    </main>
</body>
</html>`

// Update cssContent with new modern design

const cssContent = `
:root {
  --primary: #ff4757;
  --primary-light: #ff6b81;
  --secondary: #2ed573;
  --dark: #2f3542;
  --light: #f1f2f6;
  --gray: #a4b0be;
  --card-shadow: 0 10px 20px rgba(0, 0, 0, 0.1);
  --transition: all 0.3s ease;
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
  font-family: 'Poppins', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

body {
  line-height: 1.6;
  color: var(--dark);
  background-color: #f9f9f9;
}

header {
  background: linear-gradient(135deg, var(--primary), var(--primary-light));
  color: white;
  text-align: center;
  padding: 1.5rem;
  box-shadow: 0 2px 15px rgba(0, 0, 0, 0.1);
}

.logo {
  font-size: 2.2rem;
  font-weight: 700;
  letter-spacing: 1px;
}

/* Modern Navbar styling */
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background-color: white;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  position: sticky;
  top: 0;
  z-index: 1000;
}

.nav-left {
  display: flex;
  align-items: center;
}

.nav-logo {
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--primary);
  text-decoration: none;
  padding: 0;
}

.nav-links {
  display: flex;
  gap: 1rem;
}

.nav-links a {
  text-decoration: none;
  color: var(--dark);
  font-weight: 500;
  padding: 0.5rem 1rem;
  border-radius: 5px;
  transition: all 0.3s ease;
}

.nav-links a:hover {
  background-color: var(--light);
  color: var(--primary);
}

.nav-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.welcome-text {
  font-weight: 500;
  color: var(--dark);
}

.nav-btn {
  padding: 0.5rem 1.5rem;
  border-radius: 50px;
  text-decoration: none;
  font-weight: 600;
  transition: all 0.3s ease;
}

.login-btn {
  color: var(--primary);
  background-color: transparent;
  border: 1px solid var(--primary);
}

.login-btn:hover {
  background-color: var(--primary-light);
  color: white;
}

.signup-btn, .logout-btn {
  background-color: var(--primary);
  color: white;
}

.signup-btn:hover {
  background-color: var(--primary-light);
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(255, 71, 87, 0.3);
}

.logout-btn:hover {
  background-color: #ff3547;
}

/* Mobile responsive menu */
@media (max-width: 768px) {
  .navbar {
    flex-direction: column;
    padding: 1rem;
  }
  
  .nav-left, .nav-links, .nav-right {
    width: 100%;
    margin-bottom: 0.5rem;
  }
  
  .nav-links {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .nav-right {
    justify-content: center;
  }
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

h2 {
  margin-bottom: 1.5rem;
  color: var(--dark);
  position: relative;
  display: inline-block;
}

h2:after {
  content: '';
  position: absolute;
  width: 50%;
  height: 3px;
  background-color: var(--primary);
  bottom: -8px;
  left: 0;
}

.movie-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 25px;
  margin-top: 20px;
}

.movie-card {
  background-color: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: var(--card-shadow);
  transition: var(--transition);
  position: relative;
}

.movie-card:hover {
  transform: translateY(-10px);
  box-shadow: 0 15px 30px rgba(0, 0, 0, 0.15);
}

.movie-image {
  width: 100%;
  height: 350px;
  object-fit: cover;
  transition: var(--transition);
}

.movie-card:hover .movie-image {
  transform: scale(1.05);
}

.movie-details {
  padding: 20px;
}

.movie-title {
  font-size: 1.3rem;
  margin-bottom: 12px;
  color: var(--dark);
  font-weight: 600;
}

.movie-info {
  margin-bottom: 20px;
  color: #707070;
}

.movie-info span {
  display: block;
  margin-bottom: 8px;
  font-size: 0.95rem;
}

.btn {
  display: inline-block;
  background-color: var(--primary);
  color: white;
  padding: 10px 20px;
  border-radius: 30px;
  text-decoration: none;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: var(--transition);
  text-align: center;
  box-shadow: 0 4px 8px rgba(255, 71, 87, 0.2);
}

.btn:hover {
  background-color: var(--primary-light);
  transform: translateY(-2px);
  box-shadow: 0 6px 12px rgba(255, 71, 87, 0.3);
}

.btn-secondary {
  background-color: var(--secondary);
  box-shadow: 0 4px 8px rgba(46, 213, 115, 0.2);
}

.btn-secondary:hover {
  background-color: #26bd65;
  box-shadow: 0 6px 12px rgba(46, 213, 115, 0.3);
}

.btn-danger {
  background-color: #ff6b6b;
  box-shadow: 0 4px 8px rgba(255, 107, 107, 0.2);
}

.btn-danger:hover {
  background-color: #ee5253;
  box-shadow: 0 6px 12px rgba(255, 107, 107, 0.3);
}

.form {
  background-color: white;
  padding: 30px;
  border-radius: 12px;
  box-shadow: var(--card-shadow);
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #576574;
}

.form-control {
  width: 100%;
  padding: 12px 15px;
  border: 1px solid #dfe4ea;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.3s;
}

.form-control:focus {
  border-color: var(--primary);
  outline: none;
  box-shadow: 0 0 0 3px rgba(255, 71, 87, 0.1);
}

.alert {
  padding: 15px 20px;
  border-radius: 8px;
  margin-bottom: 25px;
}

.alert-danger {
  background-color: #ffe0e3;
  color: #cf000f;
  border-left: 4px solid #ff6b6b;
}

.alert-success {
  background-color: #e3ffe2;
  color: #0a8f08;
  border-left: 4px solid #2ed573;
}

/* Seating chart styles */
.seat-grid {
  display: grid;
  grid-template-columns: repeat(10, 35px);
  gap: 8px;
  margin: 25px 0;
  justify-content: center;
}

.seat {
  width: 35px;
  height: 35px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 1px solid #dfe4ea;
  cursor: pointer;
  font-size: 0.8rem;
  transition: var(--transition);
  background-color: white;
}

.seat:hover:not(.booked) {
  background-color: var(--light);
  border-color: var(--primary);
}

.seat.booked {
  background-color: #ff6b6b;
  color: white;
  cursor: not-allowed;
  border-color: #ff6b6b;
}

.seat.selected {
  background-color: var(--secondary);
  color: white;
  border-color: var(--secondary);
}

.screen {
  width: 80%;
  height: 40px;
  background: linear-gradient(0deg, #dfe4ea 0%, #f1f2f6 100%);
  margin: 0 auto 30px;
  border-radius: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #576574;
  font-weight: 600;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.1);
  transform: perspective(300px) rotateX(-5deg);
}

/* Booking list styles */
.bookings-list {
  margin-top: 25px;
}

.booking-item {
  background-color: white;
  border-radius: 12px;
  box-shadow: var(--card-shadow);
  margin-bottom: 20px;
  overflow: hidden;
  transition: var(--transition);
}

.booking-item:hover {
  transform: translateY(-5px);
  box-shadow: 0 15px 30px rgba(0, 0, 0, 0.1);
}

.booking-header {
  background-color: var(--light);
  padding: 15px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.booking-details {
  padding: 20px;
}

.booking-actions {
  padding: 0 20px 20px;
  display: flex;
  justify-content: space-between;
}

/* Admin dashboard styles */
.admin-stats {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 25px;
  margin: 25px 0;
}

.stat-card {
  background-color: white;
  border-radius: 12px;
  padding: 25px;
  box-shadow: var(--card-shadow);
  text-align: center;
  transition: var (--transition);
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 15px 30px rgba(0, 0, 0, 0.1);
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--primary);
  margin: 15px 0;
}

/* Search form styles */
.search-form {
  margin: 25px 0;
  display: flex;
  justify-content: center;
}

.search-form .form-group {
  flex: 1;
  max-width: 600px;
  margin-bottom: 0;
  position: relative;
}

.search-form .form-control {
  padding-right: 120px;
  border-radius: 30px;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.05);
}

.search-form .btn {
  position: absolute;
  right: 5px;
  top: 5px;
  height: calc(100% - 10px);
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .container {
    padding: 1rem;
  }
  
  .movie-grid {
    grid-template-columns: 1fr;
  }
  
  nav {
    flex-wrap: wrap;
  }
  
  nav a {
    margin-bottom: 5px;
  }
  
  .admin-stats {
    grid-template-columns: 1fr;
  }
  
  .seat-grid {
    grid-template-columns: repeat(10, 30px);
    gap: 5px;
  }
  
  // Around line ~1120, at the end of the cssContent constant

// Find the end of your existing CSS:
  .seat {
    width: 30px;
    height: 30px;
    font-size: 0.7rem;
  }
}

// Add the new CSS right here, before the closing backtick
.booking-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
  margin-top: 20px;
}

.movie-preview {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: var (--card-shadow);
  grid-column: 1 / 2;
}

.movie-poster {
  width: 100%;
  height: 300px;
  object-fit: cover;
}

// [Add the rest of your CSS here]

@media (max-width: 768px) {
  .booking-container {
    grid-template-columns: 1fr;
  }
  
  .movie-preview, .seat-selection, .booking-form-container {
    grid-column: 1;
  }
}
` // <-- This is the closing backtick for const cssContent

// Landing page template
const landingTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Welcome to Moobee - Premium Movie Experience</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles.css">
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        .hero {
            min-height: 90vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            text-align: center;
            background: linear-gradient(rgba(0,0,0,0.7), rgba(0,0,0,0.7)), 
                        url('/static/images/movie_1.jpg') center/cover no-repeat;
            color: white;
            padding: 2rem;
        }
        
        .hero h1 {
            font-size: 3.5rem;
            margin-bottom: 1rem;
            animation: fadeIn 1s ease-in;
        }
        
        .hero p {
            font-size: 1.5rem;
            max-width: 800px;
            margin-bottom: 2rem;
            animation: fadeIn 1.5s ease-in;
        }
        
        .cta-buttons {
            display: flex;
            gap: 20px;
            animation: fadeIn 2s ease-in;
        }
        
        .cta-btn {
            display: inline-block;
            padding: 15px 40px;
            border-radius: 50px;
            font-size: 1.2rem;
            font-weight: 600;
            text-decoration: none;
            transition: all 0.3s ease;
        }
        
        .cta-primary {
            background-color: var(--primary);
            color: white;
            box-shadow: 0 4px 15px rgba(255, 71, 87, 0.4);
        }
        
        .cta-primary:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 20px rgba(255, 71, 87, 0.5);
        }
        
        .cta-secondary {
            background-color: transparent;
            color: white;
            border: 2px solid white;
        }
        
        .cta-secondary:hover {
            background-color: white;
            color: var(--primary);
            transform: translateY(-5px);
        }
        
        .features {
            padding: 4rem 2rem;
            background-color: white;
        }
        
        .features-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 30px;
            max-width: 1200px;
            margin: 0 auto;
        }
        
        .feature-card {
            text-align: center;
            padding: 2rem;
            border-radius: 12px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.08);
            transition: all 0.3s ease;
        }
        
        .feature-card:hover {
            transform: translateY(-10px);
            box-shadow: 0 15px 30px rgba(0,0,0,0.15);
        }
        
        .feature-icon {
            font-size: 3rem;
            margin-bottom: 1rem;
            color: var(--primary);
        }
        
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
        
        @media (max-width: 768px) {
            .hero h1 {
                font-size: 2.5rem;
            }
            
            .hero p {
                font-size: 1.2rem;
            }
            
            .cta-buttons {
                flex-direction: column;
            }
        }
    </style>
</head>
<body>
    <header>
        <div class="logo">Moobee</div>
        <p>Your Ultimate Movie Experience</p>
    </header>
    
    <section class="hero">
        <h1>Experience Movies Like Never Before</h1>
        <p>Premium seats, exceptional sound, crystal-clear visuals, and a seamless booking experience. All at your fingertips.</p>
        <div class="cta-buttons">
            <a href="/login" class="cta-btn cta-primary">Login</a>
            <a href="/register" class="cta-btn cta-secondary">Create Account</a>
        </div>
    </section>
    
    <section class="features">
        <h2 style="text-align: center; margin-bottom: 3rem;">Why Choose Moobee?</h2>
        <div class="features-grid">
            <div class="feature-card">
                <div class="feature-icon">🎬</div>
                <h3>Latest Releases</h3>
                <p>Be the first to watch the newest blockbusters in premium quality.</p>
            </div>
            <div class="feature-card">
                <div class="feature-icon">🍿</div>
                <h3>Easy Booking</h3>
                <p>Select your seats, book your tickets, and enjoy the show - all in a few clicks.</p>
            </div>
            <div class="feature-card">
                <div class="feature-icon">💺</div>
                <h3>Comfortable Seating</h3>
                <p>Relax in our premium seats designed for the ultimate movie experience.</p>
            </div>
        </div>
    </section>
    
    <footer>
        <div class="container">
            <p>&copy; 2023 Moobee. All rights reserved.</p>
        </div>
    </footer>
</body>
</html>`
