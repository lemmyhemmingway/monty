// Toggle fields based on check type
document.querySelectorAll('input[name="check_type"]').forEach(radio => {
radio.addEventListener('change', function() {
const sslFields = document.querySelector('.ssl-fields');
const dnsFields = document.querySelector('.dns-fields');
const tcpFields = document.querySelector('.tcp-fields');
const httpFields = document.querySelector('.http-fields');

// Hide all fields first
sslFields.style.display = 'none';
dnsFields.style.display = 'none';
tcpFields.style.display = 'none';
    httpFields.style.display = 'none';

// Show relevant fields
    switch (this.value) {
        case 'ssl':
            sslFields.style.display = 'block';
            break;
        case 'dns':
            dnsFields.style.display = 'block';
            break;
        case 'domain':
            // No specific fields for domain
            break;
        case 'tcp':
            tcpFields.style.display = 'block';
            break;
        case 'http':
        default:
            httpFields.style.display = 'block';
            break;
    }
});
});

// Submit form as JSON
document.getElementById('endpoint-form').addEventListener('submit', function(e) {
    e.preventDefault();
    const formData = new FormData(this);
    const data = {};

    for (let [key, value] of formData) {
    if (key === 'expected_status_codes' || key === 'acceptable_tls_versions') {
    data[key] = value.split(',').map(s => s.trim()).filter(s => s).map(s => key === 'expected_status_codes' ? parseInt(s) : s);
    } else if (key === 'expected_dns_answers') {
    data[key] = [parseInt(value)]; // ExpectedDNSAnswers is an array
    } else if (key === 'interval' || key === 'timeout' || key === 'max_response_time' || key === 'min_days_valid' || key === 'tcp_port') {
    data[key] = parseInt(value);
    } else if (key === 'check_chain' || key === 'check_domain_match') {
    data[key] = this.querySelector(`input[name="${key}"]`).checked;
    } else {
            data[key] = value;
        }
    }

    fetch('/endpoints', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
    .then(response => response.json())
    .then(result => {
        if (result.error) {
            alert('Error: ' + result.error);
        } else {
            alert('Endpoint added successfully!');
            location.reload();
        }
    })
    .catch(error => {
        alert('Error adding endpoint: ' + error);
    });
});

// Tab functionality
function showTab(type) {
    // Update active tab button
    document.querySelectorAll('.tab-button').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');

    // Show/hide tab content
    const contents = document.querySelectorAll('.tab-content');
    contents.forEach(content => {
        content.style.display = 'none';
    });
    document.getElementById(type).style.display = 'block';
}

// Delete endpoint function
function deleteEndpoint(id) {
    if (!confirm('Are you sure you want to delete this endpoint? This action cannot be undone.')) {
        return;
    }

    fetch(`/endpoints/${id}`, {
        method: 'DELETE',
    })
    .then(response => response.json())
    .then(result => {
        if (result.error) {
            alert('Error: ' + result.error);
        } else {
            alert('Endpoint deleted successfully!');
            location.reload();
        }
    })
    .catch(error => {
        alert('Error deleting endpoint: ' + error);
    });
}
