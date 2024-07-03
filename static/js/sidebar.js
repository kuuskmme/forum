document.addEventListener("DOMContentLoaded", function() {
    var categoriesLink = document.querySelector(".categories-link");
    var categoriesSidebar = document.querySelector(".sidebar");
    var categoriesSidebarTrigger = document.querySelector(".categories-link")
    // Toggle categoriesSidebar on click
    categoriesLink.addEventListener("click", function() {
        categoriesSidebar.classList.toggle('expanded');
        categoriesSidebarTrigger.style.opacity = 0; // Hide the trigger when sidebar is expanded
    });

    categoriesSidebar.addEventListener("mouseleave", function() {
        categoriesSidebar.classList.toggle('expanded');
        categoriesSidebarTrigger.style.opacity = 1; // Show the trigger when sidebar is collapsed
    });
});