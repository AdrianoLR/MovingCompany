// ─── Booking Link ─────────────────────────────────────────────────────────────

function generateBookingLink() {
    const btn = document.querySelector('button[onclick="generateBookingLink()"]');
    const originalText = btn.textContent;
    btn.disabled = true;
    btn.innerHTML = '<span class="animate-spin">⏳</span> Generating...';

    fetch('/api/generate-link', { method: 'GET' })
        .then(r => r.json())
        .then(data => {
            const linkInput = document.getElementById('booking-link');
            linkInput.value = data.booking_url;
            document.getElementById('link-result').classList.remove('hidden');
            linkInput.select();
            alert('Booking link generated successfully!');
        })
        .catch(() => alert('Failed to generate booking link. Please try again.'))
        .finally(() => {
            btn.disabled = false;
            btn.textContent = originalText;
        });
}

// ─── Invoice (top-level, no booking selected) ─────────────────────────────────

function generateInvoice() {
    const btn = document.querySelector('button[onclick="generateInvoice()"]');
    const originalText = btn.textContent;
    btn.disabled = true;
    btn.textContent = 'Generating...';

    fetch('/admin/generate-invoice', { method: 'GET', headers: { 'Accept': 'application/pdf' } })
        .then(r => {
            if (!r.ok) throw new Error('Failed to generate invoice');
            return r.blob();
        })
        .then(blob => {
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'invoice.pdf';
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        })
        .catch(() => alert('Failed to generate invoice. Please try again.'))
        .finally(() => {
            btn.disabled = false;
            btn.textContent = originalText;
        });
}

// ─── Invoice (per-booking, via modal) ─────────────────────────────────────────

let currentBookingId = null;

function generateBookingInvoice(bookingId) {
    currentBookingId = bookingId;
    document.getElementById('invoiceModal').classList.remove('hidden');
    document.getElementById('totalAmount').value = '340.00';
    document.getElementById('hoursUsed').value = '2.0';
    document.getElementById('jobDescription').value = 'Flat pack delivery no required assembly';
}

function closeInvoiceModal() {
    document.getElementById('invoiceModal').classList.add('hidden');
    currentBookingId = null;
}

function setupInvoiceFormHandler() {
    const form = document.getElementById('invoiceForm');
    if (!form) return;

    form.addEventListener('submit', function(e) {
        e.preventDefault();

        const totalAmount = document.getElementById('totalAmount').value;
        const hoursUsed   = document.getElementById('hoursUsed').value;
        const jobDescription = document.getElementById('jobDescription').value;

        if (!totalAmount || !hoursUsed) {
            alert('Please fill in all required fields');
            return;
        }

        const bookingId = currentBookingId;
        closeInvoiceModal();

        const invoiceBtn = document.querySelector(`button[onclick="generateBookingInvoice('${bookingId}')"]`);
        if (invoiceBtn) {
            invoiceBtn.disabled = true;
            invoiceBtn.textContent = 'Generating...';
        }

        const params = new URLSearchParams({
            booking_id: bookingId,
            total_amount: totalAmount,
            hours_used: hoursUsed,
            job_description: jobDescription
        });

        fetch(`/admin/generate-invoice?${params}`, { method: 'GET', headers: { 'Accept': 'application/pdf' } })
            .then(r => {
                if (!r.ok) throw new Error('Failed to generate invoice');
                return r.blob();
            })
            .then(blob => {
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = `invoice-${bookingId}.pdf`;
                document.body.appendChild(a);
                a.click();
                a.remove();
                window.URL.revokeObjectURL(url);
            })
            .catch(() => alert('Failed to generate invoice. Please try again.'))
            .finally(() => {
                if (invoiceBtn) {
                    invoiceBtn.disabled = false;
                    invoiceBtn.textContent = 'Invoice';
                }
            });
    });
}

// ─── Clipboard ────────────────────────────────────────────────────────────────

function copyToClipboard(elementId) {
    const el = document.getElementById(elementId);
    el.select();
    document.execCommand('copy');
    const btn = el.nextElementSibling;
    const original = btn.textContent;
    btn.textContent = 'Copied!';
    setTimeout(() => { btn.textContent = original; }, 2000);
}

// ─── Date / timezone helpers ──────────────────────────────────────────────────

// Convert a UTC date string to a datetime-local input value in Perth time (UTC+8, no DST)
function toLocalDatetimeValue(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return '';
    const parts = new Intl.DateTimeFormat('en-CA', {
        timeZone: 'Australia/Perth',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
    }).formatToParts(date);
    const p = {};
    parts.forEach(({ type, value }) => p[type] = value);
    return `${p.year}-${p.month}-${p.day}T${p.hour}:${p.minute}`;
}

// Convert a Perth datetime-local value ("YYYY-MM-DDTHH:mm") to a UTC ISO string for the server
function perthInputToUTC(localDatetimeStr) {
    if (!localDatetimeStr) return '';
    // +08:00 is Perth's permanent offset (no DST in WA)
    const date = new Date(localDatetimeStr + ':00+08:00');
    return date.toISOString();
}

// Format a UTC date string for display in Perth time
function formatDateUTC8(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return '';
    return date.toLocaleString('en-AU', {
        timeZone: 'Australia/Perth',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
    });
}

// ─── Bookings table ───────────────────────────────────────────────────────────

let currentSort = 'name';

function getBookingsTable() { return document.getElementById('bookings-table'); }

function getCurrentBookings() {
    try { return JSON.parse(getBookingsTable().dataset.bookings || '[]'); } catch { return []; }
}

function setCurrentBookings(bookings) {
    getBookingsTable().dataset.bookings = JSON.stringify(bookings);
}

function sortBookings(bookings, sortBy) {
    const copy = [...bookings];
    if (sortBy === 'name') {
        copy.sort((a, b) => (a.customer_name || '').localeCompare(b.customer_name || ''));
    } else if (sortBy === 'status') {
        copy.sort((a, b) => (a.status - b.status) || (a.customer_name || '').localeCompare(b.customer_name || ''));
    }
    return copy;
}

function handleSortChange() {
    currentSort = document.getElementById('sort-bookings').value;
    getBookingsTable().innerHTML = formatBookingTable(getCurrentBookings());
}

function formatBookingTable(bookings) {
    bookings = sortBookings(bookings, currentSort);

    function getStatusText(status) {
        switch (status) {
            case 0: return 'Pending';
            case 1: return 'Confirmed';
            case 2: return 'In Progress';
            case 3: return 'Completed';
            case 4: return 'Cancelled';
            default: return 'Unknown';
        }
    }

    return `
        <table class="admin-table">
            <thead>
                <tr>
                    <th>Customer</th>
                    <th>Email</th>
                    <th>Phone</th>
                    <th>Pickup Address</th>
                    <th>Drop Address</th>
                    <th>Pickup Date</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${bookings.map(booking => {
                    const isEditing = booking.isEditing || false;
                    const pickupDate = new Date(booking.pickup_date);
                    const formattedDate = !isNaN(pickupDate.getTime())
                        ? toLocalDatetimeValue(booking.pickup_date)
                        : toLocalDatetimeValue(new Date().toISOString());

                    return `
                    <tr id="booking-row-${booking.user_id}" data-booking-id="${booking.user_id}" data-pickup-date="${booking.pickup_date}">
                        <td data-label="Customer">
                            ${isEditing
                                ? `<input type="text" class="w-full p-2 border rounded" value="${booking.customer_name}">`
                                : booking.customer_name}
                        </td>
                        <td data-label="Email">
                            ${isEditing
                                ? `<input type="email" class="w-full p-2 border rounded" value="${booking.email}">`
                                : booking.email}
                        </td>
                        <td data-label="Phone">
                            ${isEditing
                                ? `<input type="tel" class="w-full p-2 border rounded" value="${booking.phone}">`
                                : booking.phone}
                        </td>
                        <td data-label="Pickup Address">
                            ${isEditing
                                ? `<input type="text" class="w-full p-2 border rounded" value="${booking.pickup_address}">`
                                : booking.pickup_address}
                        </td>
                        <td data-label="Drop Address">
                            ${isEditing
                                ? `<input type="text" class="w-full p-2 border rounded" value="${booking.drop_address}">`
                                : booking.drop_address}
                        </td>
                        <td data-label="Pickup Date">
                            ${isEditing
                                ? `<input type="datetime-local" class="w-full p-2 border rounded" value="${formattedDate}">`
                                : formatDateUTC8(booking.pickup_date)}
                        </td>
                        <td data-label="Status">
                            <select class="status-select"
                                ${!isEditing ? `onchange="updateStatus('${booking.user_id}', this.value)"` : ''}
                                ${!isEditing ? 'disabled' : ''}>
                                <option value="0" ${booking.status === 0 ? 'selected' : ''}>Pending</option>
                                <option value="1" ${booking.status === 1 ? 'selected' : ''}>Confirmed</option>
                                <option value="2" ${booking.status === 2 ? 'selected' : ''}>In Progress</option>
                                <option value="3" ${booking.status === 3 ? 'selected' : ''}>Completed</option>
                                <option value="4" ${booking.status === 4 ? 'selected' : ''}>Cancelled</option>
                            </select>
                        </td>
                        <td data-label="Actions">
                            ${isEditing
                                ? `<button class="save-btn mr-2" onclick="saveBooking('${booking.user_id}')">Save</button>
                                   <button class="cancel-btn" onclick="cancelEdit('${booking.user_id}')">Cancel</button>`
                                : `<button class="edit-btn mr-2" onclick="editBooking('${booking.user_id}')">Edit</button>
                                   <button class="bg-green-600 text-white px-3 py-1 rounded text-sm hover:bg-green-700" onclick="generateBookingInvoice('${booking.user_id}')">Invoice</button>`
                            }
                        </td>
                    </tr>`;
                }).join('')}
            </tbody>
        </table>`;
}

function editBooking(id) {
    const table = getBookingsTable();
    try {
        const bookings = JSON.parse(table.dataset.bookings || '[]');
        const updated = bookings.map(b => ({ ...b, isEditing: b.user_id === id }));
        table.dataset.bookings = JSON.stringify(updated);
        table.innerHTML = formatBookingTable(updated);
    } catch (err) {
        console.error('Error editing booking:', err);
    }
}

function saveBooking(id) {
    const row = document.getElementById(`booking-row-${id}`);
    const inputs = row.getElementsByTagName('input');
    const statusSelect = row.querySelector('select');

    const updatedBooking = {
        user_id: id,
        customer_name: inputs[0].value,
        email: inputs[1].value,
        phone: inputs[2].value,
        pickup_address: inputs[3].value,
        drop_address: inputs[4].value,
        pickup_date: perthInputToUTC(inputs[5].value),
        status: parseInt(statusSelect.value, 10)
    };

    fetch(`/api/bookings/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updatedBooking)
    })
    .then(r => {
        if (!r.ok) throw new Error('Network response was not ok');
        return fetch('/api/bookings/');
    })
    .then(r => r.json())
    .then(bookings => {
        setCurrentBookings(bookings);
        getBookingsTable().innerHTML = formatBookingTable(bookings);
    })
    .catch(err => {
        console.error('Error:', err);
        alert('Error updating booking');
    });
}

function cancelEdit(id) {
    const table = getBookingsTable();
    try {
        const bookings = JSON.parse(table.dataset.bookings || '[]');
        const updated = bookings.map(b => ({ ...b, isEditing: false }));
        table.dataset.bookings = JSON.stringify(updated);
        table.innerHTML = formatBookingTable(updated);
    } catch (err) {
        console.error('Error canceling edit:', err);
    }
}

function updateStatus(id, status) {
    const row = document.getElementById(`booking-row-${id}`);
    const pickupDateIso = row.dataset.pickupDate;
    const cells = row.getElementsByTagName('td');

    const updatedBooking = {
        user_id: id,
        customer_name: cells[0].textContent.trim(),
        email: cells[1].textContent.trim(),
        phone: cells[2].textContent.trim(),
        pickup_address: cells[3].textContent.trim(),
        drop_address: cells[4].textContent.trim(),
        pickup_date: pickupDateIso,
        status: parseInt(status, 10)
    };

    fetch(`/api/bookings/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updatedBooking)
    })
    .then(r => {
        if (!r.ok) throw new Error('Network response was not ok');
        return fetch('/api/bookings/');
    })
    .then(r => r.json())
    .then(bookings => {
        setCurrentBookings(bookings);
        getBookingsTable().innerHTML = formatBookingTable(bookings);
    })
    .catch(err => {
        console.error('Error:', err);
        alert('Error updating status');
    });
}

// ─── HTMX event handlers ──────────────────────────────────────────────────────

document.body.addEventListener('htmx:afterRequest', function(evt) {
    if (evt.detail.target.id !== 'bookings-table' || !evt.detail.successful) return;
    try {
        const bookings = JSON.parse(evt.detail.xhr.response);
        if (!bookings || bookings.length === 0) {
            evt.detail.target.innerHTML = '<p class="text-gray-500 p-4">No bookings found in the database.</p>';
            return;
        }
        evt.detail.target.dataset.bookings = JSON.stringify(bookings);
        evt.detail.target.innerHTML = formatBookingTable(bookings);
    } catch (err) {
        evt.detail.target.innerHTML = `<p class="text-red-500 p-4">Error loading bookings. Response was: ${evt.detail.xhr.response}</p>`;
    }
});

document.body.addEventListener('htmx:responseError', function(evt) {
    if (evt.detail.target.id === 'bookings-table') {
        console.error('HTMX request error:', evt.detail);
        evt.detail.target.innerHTML = `<p class="text-red-500 p-4">Error loading bookings: ${evt.detail.error}</p>`;
    }
});

// ─── Init ─────────────────────────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', function() {
    const sortSelect = document.getElementById('sort-bookings');
    if (sortSelect) currentSort = sortSelect.value;
    setupInvoiceFormHandler();
});
