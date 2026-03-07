// Allow HTMX to use <template> elements when parsing responses.
// This is required for <tr> fragments (Edit/Save/Cancel row swaps) because
// browsers silently drop <tr> when inserted into a <div>, which is the
// default HTMX parsing strategy.
htmx.config.useTemplateFragments = true;

// Tracks the booking currently open in the invoice modal
let currentBookingId = null;

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
        const hoursUsed = document.getElementById('hoursUsed').value;
        const jobDescription = document.getElementById('jobDescription').value;

        if (!totalAmount || !hoursUsed) {
            alert('Please fill in all required fields');
            return;
        }

        // Capture booking ID and close modal before locating the button,
        // so closeInvoiceModal()'s null-out doesn't affect the captured value.
        const bookingId = currentBookingId;
        closeInvoiceModal();

        const params = new URLSearchParams({
            booking_id: bookingId,
            total_amount: totalAmount,
            hours_used: hoursUsed,
            job_description: jobDescription
        });

        // Find the row's Invoice button for loading-state feedback.
        // The fetch is made regardless so the invoice is always attempted.
        const invoiceBtn = document.querySelector(`button[onclick="generateBookingInvoice('${bookingId}')"]`);
        if (invoiceBtn) {
            invoiceBtn.disabled = true;
            invoiceBtn.textContent = 'Generating...';
        }

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

// ─── Bookings table sort ──────────────────────────────────────────────────────

function getBookingsTable() { return document.getElementById('bookings-table'); }

function handleSortChange() {
    const sort = document.getElementById('sort-bookings').value;
    htmx.ajax('GET', `/api/bookings/?sort=${sort}`, {
        target: getBookingsTable(),
        swap: 'innerHTML'
    });
}

// ─── Init ─────────────────────────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', function() {
    setupInvoiceFormHandler();
});
