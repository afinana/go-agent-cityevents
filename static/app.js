document.addEventListener('DOMContentLoaded', () => {
    const searchForm = document.getElementById('search-form');
    const searchInput = document.getElementById('search-input');
    const welcomeScreen = document.getElementById('welcome-screen');
    const resultsContainer = document.getElementById('results-container');
    const loadingContainer = document.getElementById('loading-container');

    searchForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const query = searchInput.value.trim();
        if (!query) return;

        // UI State: Loading
        welcomeScreen.style.display = 'none';
        resultsContainer.style.display = 'none';
        resultsContainer.innerHTML = '';
        loadingContainer.style.display = 'flex';

        try {
            const response = await fetch('/api/search', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ query })
            });

            if (!response.ok) {
                throw new Error('Search request failed');
            }

            const events = await response.json();
            
            // UI State: Results
            loadingContainer.style.display = 'none';
            renderResults(events);
            
        } catch (error) {
            console.error('Error:', error);
            loadingContainer.style.display = 'none';
            resultsContainer.innerHTML = `
                <div class="result-card" style="border-color: #ff5f5f;">
                    <h3 style="color: #ff5f5f;">Error</h3>
                    <p>Failed to retrieve events. Make sure the backend and MongoDB are running.</p>
                </div>
            `;
            resultsContainer.style.display = 'flex';
        }
    });

    function renderResults(events) {
        if (!events || events.length === 0) {
            resultsContainer.innerHTML = `
                <div class="result-card">
                    <h3>No results found</h3>
                    <p>Try rephrasing your search query.</p>
                </div>
            `;
            resultsContainer.style.display = 'flex';
            return;
        }

        // Generate HTML for each event with staggered animation
        events.forEach((event, index) => {
            const delay = index * 0.1;
            const card = document.createElement('div');
            card.className = 'result-card';
            card.style.animationDelay = `${delay}s`;
            
            // Limit description length for UI
            const desc = event.description.length > 200 
                ? event.description.substring(0, 200) + '...' 
                : event.description;

            card.innerHTML = `
                <h3>${event.title}</h3>
                <p>${desc}</p>
                <div class="result-meta">
                    <span><i class="fa-solid fa-location-dot"></i> Madrid</span>
                    <span><i class="fa-solid fa-calendar-day"></i> Upcoming</span>
                </div>
            `;
            
            resultsContainer.appendChild(card);
        });

        resultsContainer.style.display = 'flex';
    }
});
