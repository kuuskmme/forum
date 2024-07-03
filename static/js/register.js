document.addEventListener('DOMContentLoaded', () => {
    const signUpButton = document.getElementById('signUp');
    const signInButton = document.getElementById('signIn');
    const container = document.getElementById('register-container');

    // Check localStorage for the saved panel state
    const activePanel = localStorage.getItem('activePanel');

    if (activePanel === 'signUp') {
        container.classList.add("right-panel-active");
    } else {
        container.classList.remove("right-panel-active");
    }

    signUpButton.addEventListener('click', () => {
        container.classList.add("right-panel-active");
        localStorage.setItem('activePanel', 'signUp'); // Save the current state
    });

    signInButton.addEventListener('click', () => {
        container.classList.remove("right-panel-active");
        localStorage.setItem('activePanel', 'signIn'); // Save the current state
    });
});
