(function() {
    const originalFetch = window.fetch;
    window.fetch = function() {
        return originalFetch.apply(this, arguments)
            .then(response => {
                if (!response.ok) {
                    console.log('Global Handler - HTTP Error Response:', response.status);
                    if (response.status >= 500 || response.status >= 400) {
                        window.location.href = '/oops';
                    }
                    // You can insert global handling logic here based on the status
                }
                return response; // Ensure response is returned for downstream processing
            });
    };
})();