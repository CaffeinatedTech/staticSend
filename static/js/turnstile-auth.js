/**
 * Turnstile Authentication Widget Manager
 * Handles Turnstile widget initialization and re-initialization after HTMX updates
 */

class TurnstileAuthManager {
    constructor() {
        this.widgets = new Map();
        this.publicKey = null;
        this.initializeEventListeners();
    }

    /**
     * Initialize event listeners for HTMX and DOM events
     */
    initializeEventListeners() {
        // Listen for HTMX after swap events to reinitialize widgets
        document.addEventListener('htmx:afterSwap', (event) => {
            this.reinitializeWidgets(event.detail.target);
        });

        // Listen for DOM content loaded to initialize widgets on page load
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => {
                this.initializeWidgets();
            });
        } else {
            this.initializeWidgets();
        }

        // Listen for HTMX before request to get current widget response
        document.addEventListener('htmx:configRequest', (event) => {
            this.injectTurnstileResponse(event);
        });
    }

    /**
     * Initialize all Turnstile widgets on the page
     */
    initializeWidgets() {
        const containers = document.querySelectorAll('.cf-turnstile-auth');
        containers.forEach(container => {
            if (!container.dataset.initialized) {
                this.initializeWidget(container);
            }
        });
    }

    /**
     * Reinitialize widgets after HTMX content swap
     */
    reinitializeWidgets(target) {
        // Find all turnstile containers in the swapped content
        const containers = target.querySelectorAll ? 
            target.querySelectorAll('.cf-turnstile-auth') : 
            (target.classList && target.classList.contains('cf-turnstile-auth') ? [target] : []);
        
        containers.forEach(container => {
            // Remove old widget if it exists
            if (container.dataset.widgetId) {
                this.removeWidget(container.dataset.widgetId);
            }
            // Initialize new widget
            this.initializeWidget(container);
        });
    }

    /**
     * Initialize a single Turnstile widget
     */
    initializeWidget(container) {
        if (!window.turnstile) {
            console.warn('Turnstile not loaded yet, retrying in 100ms...');
            setTimeout(() => this.initializeWidget(container), 100);
            return;
        }

        const publicKey = container.dataset.sitekey;
        if (!publicKey) {
            console.warn('No Turnstile public key found in data-sitekey attribute');
            return;
        }

        try {
            const widgetId = window.turnstile.render(container, {
                sitekey: publicKey,
                callback: (token) => {
                    this.onSuccess(container, token);
                },
                'error-callback': (error) => {
                    this.onError(container, error);
                },
                'expired-callback': () => {
                    this.onExpired(container);
                },
                theme: container.dataset.theme || 'auto',
                size: container.dataset.size || 'normal'
            });

            container.dataset.widgetId = widgetId;
            container.dataset.initialized = 'true';
            this.widgets.set(widgetId, container);

        } catch (error) {
            console.error('Failed to initialize Turnstile widget:', error);
        }
    }

    /**
     * Remove a widget
     */
    removeWidget(widgetId) {
        if (window.turnstile && this.widgets.has(widgetId)) {
            try {
                window.turnstile.remove(widgetId);
            } catch (error) {
                console.warn('Failed to remove Turnstile widget:', error);
            }
            this.widgets.delete(widgetId);
        }
    }

    /**
     * Handle successful Turnstile verification
     */
    onSuccess(container, token) {
        // Store the token in a hidden input or data attribute
        let tokenInput = container.parentElement.querySelector('input[name="cf-turnstile-response"]');
        if (!tokenInput) {
            tokenInput = document.createElement('input');
            tokenInput.type = 'hidden';
            tokenInput.name = 'cf-turnstile-response';
            container.parentElement.appendChild(tokenInput);
        }
        tokenInput.value = token;

        // Enable submit button if it was disabled
        const form = container.closest('form');
        if (form) {
            const submitButton = form.querySelector('button[type="submit"]');
            if (submitButton) {
                submitButton.disabled = false;
                submitButton.classList.remove('opacity-50', 'cursor-not-allowed');
            }
        }

        // Dispatch custom event
        container.dispatchEvent(new CustomEvent('turnstile:success', {
            detail: { token }
        }));
    }

    /**
     * Handle Turnstile error
     */
    onError(container, error) {
        console.error('Turnstile error:', error);
        
        // Disable submit button
        const form = container.closest('form');
        if (form) {
            const submitButton = form.querySelector('button[type="submit"]');
            if (submitButton) {
                submitButton.disabled = true;
                submitButton.classList.add('opacity-50', 'cursor-not-allowed');
            }
        }

        // Dispatch custom event
        container.dispatchEvent(new CustomEvent('turnstile:error', {
            detail: { error }
        }));
    }

    /**
     * Handle Turnstile token expiration
     */
    onExpired(container) {
        // Clear the token
        const tokenInput = container.parentElement.querySelector('input[name="cf-turnstile-response"]');
        if (tokenInput) {
            tokenInput.value = '';
        }

        // Disable submit button
        const form = container.closest('form');
        if (form) {
            const submitButton = form.querySelector('button[type="submit"]');
            if (submitButton) {
                submitButton.disabled = true;
                submitButton.classList.add('opacity-50', 'cursor-not-allowed');
            }
        }

        // Dispatch custom event
        container.dispatchEvent(new CustomEvent('turnstile:expired'));
    }

    /**
     * Inject Turnstile response into HTMX requests
     */
    injectTurnstileResponse(event) {
        const form = event.detail.elt.closest('form');
        if (!form) return;

        const turnstileContainer = form.querySelector('.cf-turnstile-auth');
        if (!turnstileContainer) return;

        const tokenInput = form.querySelector('input[name="cf-turnstile-response"]');
        if (tokenInput && tokenInput.value) {
            // Token is already in the form, HTMX will include it automatically
            return;
        }

        // If no token is present, prevent the request
        event.preventDefault();
        console.warn('Turnstile token missing, preventing form submission');
    }

    /**
     * Reset a specific widget
     */
    resetWidget(container) {
        const widgetId = container.dataset.widgetId;
        if (widgetId && window.turnstile) {
            try {
                window.turnstile.reset(widgetId);
            } catch (error) {
                console.warn('Failed to reset Turnstile widget:', error);
            }
        }
    }

    /**
     * Reset all widgets
     */
    resetAllWidgets() {
        this.widgets.forEach((container, widgetId) => {
            this.resetWidget(container);
        });
    }
}

// Initialize the manager when the script loads
const turnstileAuthManager = new TurnstileAuthManager();

// Export for global access if needed
window.turnstileAuthManager = turnstileAuthManager;
