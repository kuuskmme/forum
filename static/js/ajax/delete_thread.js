document.addEventListener("DOMContentLoaded", function() {
    // Assuming there might be multiple delete buttons in the future, let's use event delegation
    document.body.addEventListener('click', function(e) {
        if (e.target && e.target.id === 'delete-thread-btn') {
            const threadid = e.target.getAttribute('data-thread-id');
            fetch('/thread/delete-thread', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ threadID: threadid }),
            })
            .then(response => response.json())
            .then(data => {
                console.log('Success:', data.message);
                window.location.reload();
                // Optionally, remove the thread from the DOM or refresh the page
            })
            .catch(error => console.error('Error:', error));
        }
    });
});
