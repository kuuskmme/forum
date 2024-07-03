document.addEventListener("DOMContentLoaded", function() {
    // Assuming there might be multiple delete buttons in the future, let's use event delegation
    document.body.addEventListener('click', function(e) {
        if (e.target && e.target.id === 'delete-post-btn') {
            const postid = e.target.getAttribute('data-post-id');
            fetch('/thread/delete-post', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ postID: postid }),
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
