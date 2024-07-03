document.addEventListener('DOMContentLoaded', function() {
    function toggleChange(field) {
        const changeRequestedField = document.getElementById(field + "ChangeRequested");
        const newValueField = document.getElementById("new" + field);
        const changeRequested = changeRequestedField.value === "false";
    
        changeRequestedField.value = changeRequested ? "true" : "false";
    
        // Show or hide the new value field based on whether a change is requested
        newValueField.style.display = changeRequested ? "flex" : "none";
    
        // Specifically for the password field, add or remove the required attribute based on visibility
        if (field === "password") {
            if (changeRequested) {
                newValueField.setAttribute('required', '');
            } else {
                newValueField.removeAttribute('required');
            }
        }
    }
    
    // Attach event listeners
    document.getElementById("changeEmailBtn").addEventListener("click", function() {
        toggleChange("email");
    });
    document.getElementById("changePasswordBtn").addEventListener("click", function() {
        toggleChange("password");
    });
});