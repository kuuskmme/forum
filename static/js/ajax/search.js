document.getElementById('searchBtn').addEventListener('click', function() {
    var searchText = document.getElementById('searchInput').value;
    var searchType = document.getElementById('searchType').value;
    
    if (searchText.trim() && searchType) {
        var searchUrl = `/search/${searchType}/${encodeURIComponent(searchText)}`;
        
        fetch(searchUrl)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                console.log(data); // Check the structure of the response
                // Make sure to access the threads using the correct property
                    updateThreadList(data.threads);
            })
            .catch(error => {
                console.error('Failed to fetch search results:', error);
            });
    } else {
        alert('Please select a search type and enter search text.');
    }
});

function updateThreadList(threads) {
    const container = document.querySelector('.threads-container');
    // Clear existing threads
    container.innerHTML = '';
    console.log(threads)
    // Check if there are threads
    if (threads && threads.length > 0) {
        const ul = document.createElement('ul');
        ul.className = 'threads-list';

        threads.forEach(thread => {
            const li = document.createElement('li');
            li.className = 'thread-item';
            li.innerHTML = `<a href="/thread/${thread.ID}">${thread.Category.Name} > ${thread.Topic}</a>
                            <div class="thread-footer">
                                <p class="thread-date-stamp">Created ${thread.CreatedAt}</p>
                            </div>`;
            ul.appendChild(li);
        });

        container.appendChild(ul);
    } else {
        console.log("No threads found!")
        // If no threads were found, show a message
        container.innerHTML = `<section class="entries-none">
                                    <h1 class="threads-not-found">No threads match your search criteria.</h1>
                                </section>`;
    }
}
