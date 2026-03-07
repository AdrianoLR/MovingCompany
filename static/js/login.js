document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('login-form');
    const errorMessage = document.getElementById('error-message');

    loginForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;

        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });

            if (response.ok) {
                // Cookie is set server-side via HttpOnly
                window.location.href = '/admin';
            } else {
                errorMessage.style.display = 'block';
                console.error('Login error: authentication failed', response.status);
            }
        } catch (err) {
            errorMessage.style.display = 'block';
            console.error('Login error:', err);
        }
    });
});
