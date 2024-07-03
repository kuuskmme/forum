document.addEventListener("DOMContentLoaded", function() {
    var newCategoryLink = document.getElementById('new-category-link');
    var newCategoryPopup = document.getElementById('new-category-popup');
    var newCategoryBtn = document.getElementById('new-category-btn');
    var removeCategoryBtn = document.getElementById('remove-category-btn');
    var categoryNameInput = document.getElementById('new-category-name');
    var categoryCheckResult = document.getElementById('category-check-result');

    newCategoryLink.addEventListener('click', function() {
        newCategoryPopup.style.display = 'flex'; // Show the popup
        newCategoryPopup.style.flexDirection = 'column'
        this.style.display = 'none';
    });

    newCategoryBtn.addEventListener('click', function(e) {
        e.preventDefault(); // Prevent default button action
        var categoryName = categoryNameInput.value.trim();
        if (!categoryName) {
            alert("Please enter a category name.");
            return;
        }

        // Prepare the request URL and parameters
        var requestURL = '/new-category'; // Adjust the URL to your API endpoint
        var requestData = JSON.stringify({ categoryName: categoryName });

        // Perform the AJAX request
        console.log("Sending request to:", requestURL, "with data:", requestData);
        fetch(requestURL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: requestData
        })
        .then(response => response.json())
        .then(data => {
            if (data.exists) {
                console.log("Category exists! Display error.")
                // If category exists, show error
                categoryCheckResult.textContent = data.error;
            } else {
                console.log("Category doesnt exist, we have added it. Refresh page!")
                // If category does not exist, refresh the page or redirect
                window.location.reload();
            }
        })
        .catch(error => {
            console.error('Error adding category:', error);
            categoryCheckResult.textContent = "Error adding category.";
        });
    });
    removeCategoryBtn.addEventListener('click', function(e) {
        e.preventDefault(); // Prevent default button action
        var categoryName = categoryNameInput.value.trim();
        if (!categoryName) {
            alert("Please enter a category name.");
            return;
        }

        // Prepare the request URL and parameters
        var requestURL = '/remove-category'; // Adjust the URL to your API endpoint
        var requestData = JSON.stringify({ categoryName: categoryName });

        // Perform the AJAX request
        fetch(requestURL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: requestData
        })
        .then(response => response.json())
        .then(data => {
            if (data.exists) {
                console.log("Category does not exist. Refresh!")
                // If category exists, show error
                categoryCheckResult.textContent = data.error;
            } else {
                console.log("Category did exist and has been removed, refresh page!")
                // If category does not exist, refresh the page or redirect
                window.location.reload();
            }
        })
        .catch(error => {
            console.error('Error checking category:', error);
            categoryCheckResult.textContent = "Error checking category.";
        });
    });
});
