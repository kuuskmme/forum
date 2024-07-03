document.addEventListener("DOMContentLoaded", function() {
    // Handle form submission
    document.getElementById('reply_form').addEventListener('submit', function(e) {
        e.preventDefault(); // Prevent default form submission
        var userUUID = UserCtx.UUID;
        const threadID = document.getElementById('thread_id').value; 
        const formData = {
            body: document.getElementById('reply_content').value,
            useruuid: userUUID,
            threadid: threadID,
        };

        fetch('/thread/new-post', { // Adjust the URL to your endpoint for creating a new thread
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData),
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            window.location.reload();
            return response.json();
        })
        .then(data => {
            console.log('Success:', data);
            // Handle success (e.g., clear form, show success message, etc.)
        })
        .catch((error) => {
            console.error('Error:', error);
            // Handle failure (e.g., show error message)
        });
    });
});
