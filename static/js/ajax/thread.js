document.addEventListener("DOMContentLoaded", function () {
    updateViews(); // Track thread page views via AJAX!
    var likeButtons = document.querySelectorAll('.like-btn');
    var dislikeButtons = document.querySelectorAll('.dislike-btn');
    let currentlyHighlightedButton = null; // Track the currently highlighted button
    function toggleHighlight(button) {
        if (currentlyHighlightedButton === button) {
            // Unhighlight if clicking the already highlighted button
            button.classList.remove("highlighted");
            currentlyHighlightedButton = null;
        } else {
            // Highlight the new button and unhighlight the previous one (if any)
            if (currentlyHighlightedButton) {
                currentlyHighlightedButton.classList.remove("highlighted");
            }
            button.classList.add("highlighted");
            currentlyHighlightedButton = button;
        }
    }
    likeButtons.forEach(function (btn) {
        btn.addEventListener('click', function () {
            updateRating(this, true);
            toggleHighlight(this);
        });
    });

    dislikeButtons.forEach(function (btn) {
        btn.addEventListener('click', function () {
            updateRating(this, false);
            toggleHighlight(this);
        });
    });

    function getUserID(userUUID) {
        console.log("Starting to fetch user ID for UUID:", userUUID); // Initial logging

        return fetch(`/get-user-id?uuid=${encodeURIComponent(userUUID)}`)
            .then(response => {
                console.log("Received response from /get-user-id:", response); // Log the raw response

                if (!response.ok) {
                    console.error("Response not OK. Status:", response.status); // Log error if response is not OK
                    throw new Error('Network response was not ok');
                }

                return response.json(); // Attempt to parse JSON
            })
            .then(data => {
                console.log("Parsed JSON response:", data); // Log the parsed JSON data

                if (!data.userID) {
                    console.error("No userID found in response data"); // Log if userID is missing
                }

                return data.userID; // Return the userID
            })
            .catch(error => {
                console.error('Failed to fetch userID:', error); // Log any errors that occur during the fetch
            });
    }

    // Updates thread views counter
    function updateViews() {
        var userUUID = UserCtx.UUID; // Assuming UserCtx.UUID is accessible globally
        getUserID(userUUID).then(userID => {
            if (!userID) {
                console.error('Failed to retrieve userID');
                return; // Exit if no userID found
            } else {
                var threadID = window.location.pathname.split('/').pop()
                console.log(userUUID);
            }

            // Assuming your AJAX endpoint is "/update-views" and it accepts POST requests
            fetch('/thread/update-views', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    threadid: window.location.pathname.split('/').pop(), // Assuming thread ID is in the URL
                    useruuid: UserCtx.UUID
                }
                )
            })
                .then(response => response.json())  // Convert response to JSON
                .then(data => {
                    if (data.newViews !== undefined) {
                        var viewElement = document.getElementById(`thread-views-${threadID}`);
                        if (viewElement) {
                            viewElement.innerText = data.newViews; // Update the views count in the div
                        } else {
                            console.error('Element to update views count not found');
                        }
                    } else {
                        console.error('New views data is not provided in the response');
                    }
                })
                .catch(error => {
                    console.error("Also this user may have already viewed this thread if this errors btw.")
                    console.error('Error updating views:', error);
                });
        });
    }

    function updateRating(element, isLike) {
        var entityID = element.getAttribute('data-id');
        var entityType = element.getAttribute('data-type'); // 'thread' or 'post'
        var userUUID = UserCtx.UUID; // Assuming UserCtx.UUID is accessible globally
        var action = isLike ? 'like' : 'dislike';

        console.log("Initiating updateRating for", entityType, { entityID, action, userUUID });

        // Fetch userID for UUID
        getUserID(userUUID).then(userID => {
            if (!userID) {
                console.error('Failed to retrieve userID');
                return; // Exit if no userID found
            }

            // Prepare data for the POST request
            const postData = { id: entityID, type: entityType, action: action, userID: userID };
            console.log("Sending update rating request with data:", postData);

            // AJAX request for updating the rating
            var xhr = new XMLHttpRequest();
            xhr.open('POST', '/update-rating', true);
            xhr.setRequestHeader('Content-Type', 'application/json');

            xhr.onload = function () {
                if (this.status == 200) {
                    var response = JSON.parse(this.responseText);
                    console.log("Update rating successful, new rating:", response.newRating);

                    // Update the UI with the new rating
                    document.getElementById(entityType + '-rating-' + entityID).innerText = response.newRating;
                } else {
                    console.error('Error updating rating, status:', this.status);
                }
            };

            xhr.onerror = function () {
                console.error("Network error occurred while updating rating.");
            };

            xhr.send(JSON.stringify(postData));
        })
            .catch(error => {
                console.error("Error in updateRating function:", error);
            });
    }
});