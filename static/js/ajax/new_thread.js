document.addEventListener("DOMContentLoaded", function() {
    // Populate categories
    fetch('/get-categories') // Adjust the URL to your endpoint for fetching categories
        .then(response => response.json())
        .then(data => {
            console.log(data)
            const selector = document.getElementById('category-selector');
            data.categories.forEach(category => {
                let option = new Option(category.Name, category.ID);
                selector.add(option);
            });
        })
        .catch(error => console.error('Failed to load categories:', error));

    // Handle form submission
    document.getElementById('new-thread-form').addEventListener('submit', function(e) {
        e.preventDefault(); // Prevent default form submission
        const userUUID = document.getElementById('user-uuid').value; // Example
        const formData = {
            category: document.getElementById('category-selector').value,
            topic: document.getElementById('topic-field').value,
            body: document.getElementById('text-body-field').value,
            user_uuid: userUUID,
        };

        fetch('/thread/new-thread', { // Adjust the URL to your endpoint for creating a new thread
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
