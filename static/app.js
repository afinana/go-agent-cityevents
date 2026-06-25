document.addEventListener('DOMContentLoaded', () => {
    const searchForm = document.getElementById('search-form');
    const searchInput = document.getElementById('search-input');
    const welcomeScreen = document.getElementById('welcome-screen');
    const loadingContainer = document.getElementById('loading-container');
    const chatMessagesContainer = document.getElementById('chat-messages-container');
    const conversationsList = document.getElementById('conversations-list');
    const btnNewChat = document.getElementById('btn-new-chat');
    const contentArea = document.getElementById('content-area');

    // MongoDB status elements
    const dbStatusBtn = document.getElementById('db-status-btn');
    const dbStatusText = document.getElementById('db-status-text');
    const dbStatusDot = document.getElementById('db-status-dot');
    const dbStatusPing = document.getElementById('db-status-ping');

    // State Variables
    let conversations = [];
    let activeConversationId = null;

    const CONVERSATIONS_STORAGE_KEY = 'cityevents_conversations_v1';

    // Check MongoDB status
    async function checkDBStatus() {
        dbStatusText.textContent = 'MongoDB: Checking...';
        dbStatusBtn.className = 'flex items-center gap-2.5 px-3 py-1.5 rounded-full border border-amber-500/30 text-[11px] bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 transition-all duration-200';
        dbStatusDot.className = 'relative inline-flex rounded-full h-2 w-2 bg-amber-500';
        dbStatusPing.className = 'absolute inline-flex h-full w-full rounded-full bg-amber-500 opacity-75 animate-ping';
        
        try {
            const response = await fetch('/api/db/status');
            const data = await response.json();
            
            if (data.connected) {
                dbStatusText.textContent = 'MongoDB: Connected';
                dbStatusBtn.className = 'flex items-center gap-2.5 px-3 py-1.5 rounded-full border border-emerald-500/30 text-[11px] bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 transition-all duration-200';
                dbStatusDot.className = 'relative inline-flex rounded-full h-2 w-2 bg-emerald-500';
                dbStatusPing.className = 'absolute inline-flex h-full w-full rounded-full bg-emerald-500 opacity-0';
            } else {
                dbStatusText.textContent = 'MongoDB: Disconnected';
                dbStatusBtn.className = 'flex items-center gap-2.5 px-3 py-1.5 rounded-full border border-rose-500/30 text-[11px] bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 transition-all duration-200';
                dbStatusDot.className = 'relative inline-flex rounded-full h-2 w-2 bg-rose-500';
                dbStatusPing.className = 'absolute inline-flex h-full w-full rounded-full bg-rose-500 opacity-75 animate-pulse';
            }
        } catch (error) {
            console.error('Error checking MongoDB status:', error);
            dbStatusText.textContent = 'MongoDB: Disconnected';
            dbStatusBtn.className = 'flex items-center gap-2.5 px-3 py-1.5 rounded-full border border-rose-500/30 text-[11px] bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 transition-all duration-200';
            dbStatusDot.className = 'relative inline-flex rounded-full h-2 w-2 bg-rose-500';
            dbStatusPing.className = 'absolute inline-flex h-full w-full rounded-full bg-rose-500 opacity-75 animate-pulse';
        }
    }

    // Add listener to DB Status Button
    dbStatusBtn.addEventListener('click', checkDBStatus);
    checkDBStatus();

    // ----------------------------------------------------
    // Conversation State & Storage functions
    // ----------------------------------------------------
    function loadConversations() {
        const data = localStorage.getItem(CONVERSATIONS_STORAGE_KEY);
        try {
            conversations = data ? JSON.parse(data) : [];
        } catch (e) {
            console.error('Error loading conversations:', e);
            conversations = [];
        }
    }

    function saveConversations() {
        localStorage.setItem(CONVERSATIONS_STORAGE_KEY, JSON.stringify(conversations));
    }

    function createConversation(firstQuery = null) {
        const id = 'conv-' + Date.now();
        const title = firstQuery 
            ? (firstQuery.length > 25 ? firstQuery.substring(0, 25) + '...' : firstQuery) 
            : 'New conversation';
            
        const newConv = {
            id: id,
            title: title,
            timestamp: Date.now(),
            messages: []
        };
        
        conversations.unshift(newConv);
        saveConversations();
        return id;
    }

    function deleteConversation(id) {
        conversations = conversations.filter(c => c.id !== id);
        saveConversations();
        
        if (activeConversationId === id) {
            if (conversations.length > 0) {
                selectConversation(conversations[0].id);
            } else {
                activeConversationId = null;
                renderConversationsList();
                showEmptyState();
            }
        } else {
            renderConversationsList();
        }
    }

    function startNewConversation() {
        const newId = createConversation();
        selectConversation(newId);
    }

    function selectConversation(id) {
        activeConversationId = id;
        renderConversationsList();
        
        const conv = conversations.find(c => c.id === id);
        if (!conv) return;

        chatMessagesContainer.innerHTML = '';
        
        if (conv.messages.length === 0) {
            showEmptyState();
        } else {
            welcomeScreen.style.display = 'none';
            chatMessagesContainer.style.display = 'flex';
            
            conv.messages.forEach(msg => {
                appendMessageBubble(msg);
            });
            
            scrollToBottom();
        }
    }

    function showEmptyState() {
        welcomeScreen.style.display = 'block';
        chatMessagesContainer.style.display = 'none';
        chatMessagesContainer.innerHTML = '';
    }

    function scrollToBottom() {
        contentArea.scrollTop = contentArea.scrollHeight;
    }

    // Render Recent Conversations in the left pane
    function renderConversationsList() {
        if (conversations.length === 0) {
            conversationsList.innerHTML = `
                <div class="text-[11px] text-text-secondary text-center py-6 px-4 select-none">
                    No recent chats. Start a new conversation above!
                </div>
            `;
            return;
        }

        conversationsList.innerHTML = '';
        conversations.forEach(conv => {
            const isActive = conv.id === activeConversationId;
            const item = document.createElement('div');
            item.className = `conversation-item group flex items-center justify-between px-3.5 py-2.5 rounded-md cursor-pointer transition-all duration-200 hover:bg-white/5 text-sm ${
                isActive ? 'bg-accent/15 text-accent font-medium' : 'text-text-secondary'
            }`;
            item.setAttribute('data-id', conv.id);

            item.innerHTML = `
                <div class="flex items-center gap-3 overflow-hidden flex-1 select-none">
                    <i class="fa-regular fa-message text-xs opacity-75"></i>
                    <span class="truncate block pr-2">${escapeHtml(conv.title)}</span>
                </div>
                <button class="delete-thread-btn bg-transparent border-none text-text-secondary hover:text-[#ff5f5f] hover:bg-[#ff5f5f]/10 p-1 rounded transition-colors duration-200" data-id="${conv.id}" title="Delete conversation">
                    <i class="fa-regular fa-trash-can text-xs"></i>
                </button>
            `;

            // Click listener for selecting conversation
            item.addEventListener('click', (e) => {
                // If clicked the delete button, do not select
                if (e.target.closest('.delete-thread-btn')) {
                    return;
                }
                selectConversation(conv.id);
            });

            // Click listener for deleting conversation
            const delBtn = item.querySelector('.delete-thread-btn');
            delBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                deleteConversation(conv.id);
            });

            conversationsList.appendChild(item);
        });
    }

    // Helper to format event cards HTML
    function renderEventsHtml(events) {
        if (!events || events.length === 0) {
            return `
                <div class="bg-bg-card border border-border-color rounded-xl p-5 shadow-md w-full">
                    <h3 class="text-base font-medium mb-2 text-white">No results found</h3>
                    <p class="text-sm text-text-secondary leading-relaxed">Try rephrasing your search query.</p>
                </div>
            `;
        }

        let html = '';
        events.forEach((event, index) => {
            const desc = event.description.length > 200 
                ? event.description.substring(0, 200) + '...' 
                : event.description;

            html += `
                <div class="bg-bg-card border border-border-color rounded-xl p-5 shadow-md w-full hover:border-accent/50 transition-all duration-300">
                    <h3 class="text-base font-medium mb-2 text-white">${escapeHtml(event.title)}</h3>
                    <p class="text-sm text-text-secondary leading-relaxed">${escapeHtml(desc)}</p>
                    <div class="mt-4 flex gap-4 text-xs text-accent justify-between items-center w-full">
                        <div class="flex gap-4">
                            <span><i class="fa-solid fa-location-dot"></i> Madrid</span>
                            <span><i class="fa-solid fa-calendar-day"></i> Upcoming</span>
                        </div>
                        <a href="${event.sourceId}" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline flex items-center gap-1.5 font-medium select-none">
                            <i class="fa-solid fa-arrow-up-right-from-square text-[10px]"></i> View Source
                        </a>
                    </div>
                </div>
            `;
        });
        return html;
    }

    // Append message bubble directly in DOM
    function appendMessageBubble(msg) {
        const bubble = document.createElement('div');
        
        if (msg.role === 'user') {
            bubble.className = 'chat-message flex gap-4 w-full justify-end';
            bubble.innerHTML = `
                <div class="flex flex-col items-end max-w-[80%]">
                    <div class="flex items-center gap-2 mb-1 text-[11px] text-text-secondary select-none">
                        <span>You</span>
                    </div>
                    <div class="bg-bg-input border border-white/5 text-text-primary px-4 py-2.5 rounded-2xl rounded-tr-none shadow-md text-sm">
                        ${escapeHtml(msg.content)}
                    </div>
                </div>
                <div class="w-8 h-8 rounded-full bg-accent/20 border border-accent/30 flex items-center justify-center text-accent text-sm self-start select-none">
                    <i class="fa-regular fa-user"></i>
                </div>
            `;
        } else { // assistant
            bubble.className = 'chat-message flex gap-4 w-full';
            
            let contentHtml = '';
            if (msg.error) {
                contentHtml = `
                    <div class="bg-bg-card border border-rose-500 rounded-xl p-5 shadow-md w-full">
                        <h3 class="text-base font-medium mb-2 text-rose-500">Error</h3>
                        <p class="text-sm text-text-secondary leading-relaxed">${escapeHtml(msg.error)}</p>
                    </div>
                `;
            } else {
                contentHtml = renderEventsHtml(msg.events);
            }

            bubble.innerHTML = `
                <div class="w-8 h-8 rounded-full bg-accent flex items-center justify-center text-white text-sm self-start select-none">
                    <i class="fa-solid fa-wand-magic-sparkles"></i>
                </div>
                <div class="flex flex-col max-w-[85%] flex-1">
                    <div class="flex items-center gap-2 mb-1 text-[11px] text-text-secondary select-none">
                        <span>Assistant</span>
                    </div>
                    <div class="w-full flex flex-col gap-4">
                        ${contentHtml}
                    </div>
                </div>
            `;
        }
        
        chatMessagesContainer.appendChild(bubble);
    }

    // Run search
    async function executeSearch(query) {
        searchInput.value = '';
        
        // Ensure there is an active conversation, otherwise create one
        if (!activeConversationId) {
            activeConversationId = createConversation(query);
        } else {
            // If the active conversation was empty, update its title to this query
            const activeConv = conversations.find(c => c.id === activeConversationId);
            if (activeConv && activeConv.messages.length === 0) {
                activeConv.title = query.length > 25 ? query.substring(0, 25) + '...' : query;
            }
        }

        const activeConv = conversations.find(c => c.id === activeConversationId);
        if (!activeConv) return;

        // Push User Message
        const userMsg = { role: 'user', content: query };
        activeConv.messages.push(userMsg);
        saveConversations();
        renderConversationsList();

        // Render immediately
        welcomeScreen.style.display = 'none';
        chatMessagesContainer.style.display = 'flex';
        appendMessageBubble(userMsg);
        scrollToBottom();

        // Loading state
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
                const errorText = await response.text();
                throw new Error(errorText || 'Search request failed');
            }

            const events = await response.json();
            
            // Push Assistant Message
            const assistantMsg = { role: 'assistant', events: events };
            activeConv.messages.push(assistantMsg);
            saveConversations();
            
            loadingContainer.style.display = 'none';
            appendMessageBubble(assistantMsg);
            scrollToBottom();
            
        } catch (error) {
            console.error('Error:', error);
            loadingContainer.style.display = 'none';
            
            const errorMsg = { role: 'assistant', error: error.message };
            activeConv.messages.push(errorMsg);
            saveConversations();
            
            appendMessageBubble(errorMsg);
            scrollToBottom();
        }
    }

    // Helper to escape HTML tags to prevent XSS
    function escapeHtml(text) {
        if (!text) return '';
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#039;'
        };
        return text.toString().replace(/[&<>"']/g, function(m) { return map[m]; });
    }

    // New Chat Button Click Listener
    btnNewChat.addEventListener('click', startNewConversation);

    // Form submit listener
    searchForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const query = searchInput.value.trim();
        if (!query) return;
        executeSearch(query);
    });

    // ----------------------------------------------------
    // App Initialization
    // ----------------------------------------------------
    loadConversations();
    renderConversationsList();
    
    // Select first conversation if it exists, otherwise show welcome screen
    if (conversations.length > 0) {
        selectConversation(conversations[0].id);
    } else {
        showEmptyState();
    }
});
