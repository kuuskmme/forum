document.addEventListener("DOMContentLoaded", function() {
    var userSidebar = document.querySelector(".user-sidebar");
    var userSidebarTrigger = document.querySelector(".user-sidebar-trigger");
    
    userSidebarTrigger.addEventListener("mouseenter", function() {
        userSidebar.classList.add('expanded');
        userSidebarTrigger.style.opacity = 0; // Hide the trigger when sidebar is expanded
    });

    userSidebar.addEventListener("mouseleave", function() {
        userSidebar.classList.remove('expanded');
        userSidebarTrigger.style.opacity = 1; // Show the trigger when sidebar is collapsed
    });
});