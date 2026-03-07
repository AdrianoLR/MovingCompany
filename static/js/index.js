document.getElementById('bookingForm').addEventListener('submit', function(e) {
    e.preventDefault();

    const form = document.getElementById('bookingForm');
    Array.from(form.elements).forEach(el => el.disabled = true);

    const pickupDate = new Date(document.getElementById('pickup_date').value);

    // Format to "YYYY-MM-DDThh:mm:ss" as expected by the server
    const formatDate = (date) =>
        date.getFullYear() + '-' +
        String(date.getMonth() + 1).padStart(2, '0') + '-' +
        String(date.getDate()).padStart(2, '0') + 'T' +
        String(date.getHours()).padStart(2, '0') + ':' +
        String(date.getMinutes()).padStart(2, '0') + ':' +
        String(date.getSeconds()).padStart(2, '0');

    const formData = {
        customer_name: document.getElementById('customer_name').value,
        email: document.getElementById('email').value,
        phone: document.getElementById('phone').value,
        pickup_address: document.getElementById('pickup_address').value,
        drop_address: document.getElementById('drop_address').value,
        pickup_date: formatDate(pickupDate),
        furniture_items: {
            chairs: parseInt(document.getElementById('chairs').value),
            table_2_seats: parseInt(document.getElementById('table_2_seats').value),
            table_3_seats: parseInt(document.getElementById('table_3_seats').value),
            table_4_plus_seats: parseInt(document.getElementById('table_4_plus_seats').value),
            fridges: parseInt(document.getElementById('fridges').value),
            washing_machines: parseInt(document.getElementById('washing_machines').value),
            dryers: parseInt(document.getElementById('dryers').value),
            dishwashers: parseInt(document.getElementById('dishwashers').value),
            boxes: parseInt(document.getElementById('boxes').value),
            pot_plants: parseInt(document.getElementById('pot_plants').value),
            mattresses: parseInt(document.getElementById('mattresses').value),
            bed_frames: parseInt(document.getElementById('bed_frames').value),
            sofas: parseInt(document.getElementById('sofas').value)
        },
        token: document.getElementById('token').value
    };

    fetch('/api/submit-booking', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
    })
    .then(async r => {
        if (!r.ok) {
            const errorText = await r.text();
            throw new Error(`Server error: ${r.status} - ${errorText}`);
        }
        return r.json();
    })
    .then(() => {
        const notification = document.getElementById('notification');
        notification.style.display = 'block';
        notification.className = 'notification success';
        notification.textContent = 'Booking created successfully!';

        Array.from(form.elements).forEach(el => el.disabled = true);
        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) submitBtn.style.display = 'none';

        const msg = document.createElement('div');
        msg.className = 'text-center mt-4 text-lg font-bold text-green-600';
        msg.textContent = 'Your booking has been submitted. Thank you!';
        form.appendChild(msg);
    })
    .catch(error => {
        const notification = document.getElementById('notification');
        notification.style.display = 'block';
        notification.className = 'notification error';
        notification.textContent = `Error creating booking: ${error.message}`;

        Array.from(form.elements).forEach(el => el.disabled = true);
        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) submitBtn.style.display = 'none';

        const msg = document.createElement('div');
        msg.className = 'text-center mt-4 text-lg font-bold text-red-600';
        msg.textContent = 'There was an error submitting your booking. Please try again later.';
        form.appendChild(msg);
    });
});
