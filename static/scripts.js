class CommentTree {
    constructor() {
        this.apiUrl = '/comments';
        this.currentPage = 1;
        this.limit = 10;
        this.currentSort = 'asc';
        this.isSearchMode = false;
        this.currentQuery = '';
        this.replyToId = null;
        this.hasMore = true;
        this.collapsedComments = new Set();

        this.initElements();
        this.attachEventListeners();
        this.loadComments();
    }

    initElements() {
        // Search elements
        this.searchInput = document.getElementById('searchInput');
        this.searchBtn = document.getElementById('searchBtn');
        this.clearBtn = document.getElementById('clearBtn');

        // Form elements
        this.authorInput = document.getElementById('authorInput');
        this.contentInput = document.getElementById('contentInput');
        this.addCommentBtn = document.getElementById('addCommentBtn');

        // Container elements
        this.commentsContainer = document.getElementById('commentsContainer');
        this.sortSelect = document.getElementById('sortSelect');
        this.loading = document.getElementById('loading');
        this.noComments = document.getElementById('noComments');

        // Pagination elements
        this.prevBtn = document.getElementById('prevBtn');
        this.nextBtn = document.getElementById('nextBtn');
        this.pageInfo = document.getElementById('pageInfo');

        // Modal elements
        this.replyModal = document.getElementById('replyModal');
        this.replyTo = document.getElementById('replyTo');
        this.replyAuthorInput = document.getElementById('replyAuthorInput');
        this.replyContentInput = document.getElementById('replyContentInput');
        this.submitReplyBtn = document.getElementById('submitReplyBtn');
        this.cancelReplyBtn = document.getElementById('cancelReplyBtn');
        this.closeModal = document.getElementById('closeModal');
    }

    attachEventListeners() {
        // Search events
        this.searchBtn.addEventListener('click', () => this.search());
        this.clearBtn.addEventListener('click', () => this.clearSearch());
        this.searchInput.addEventListener('keypress', e => {
            if (e.key === 'Enter') this.search();
        });

        // Form events
        this.addCommentBtn.addEventListener('click', () => this.createComment());
        this.contentInput.addEventListener('keypress', e => {
            if (e.key === 'Enter' && e.ctrlKey) this.createComment();
        });

        // Sort events
        this.sortSelect.addEventListener('change', () => {
            this.currentSort = this.sortSelect.value;
            this.currentPage = 1;
            this.loadComments();
        });

        // Pagination events
        this.prevBtn.addEventListener('click', () => this.prevPage());
        this.nextBtn.addEventListener('click', () => this.nextPage());

        // Modal events
        this.submitReplyBtn.addEventListener('click', () => this.submitReply());
        this.cancelReplyBtn.addEventListener('click', () => this.closeReplyModal());
        this.closeModal.addEventListener('click', () => this.closeReplyModal());
        this.replyModal.addEventListener('click', e => {
            if (e.target === this.replyModal) this.closeReplyModal();
        });
        document.addEventListener('keydown', e => {
            if (e.key === 'Escape' && this.replyModal.style.display === 'block') {
                this.closeReplyModal();
            }
        });
    }

    async apiCall(url, options = {}) {
        try {
            const response = await fetch(url, {
                headers: { 'Content-Type': 'application/json' },
                ...options
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Ошибка сервера');
            }

            return response.status === 204 ? null : await response.json();
        } catch (error) {
            console.error('API Error:', error);
            this.showError(error.message);
            throw error;
        }
    }

    showError(message) {
        // Simple error display - you could make this more sophisticated
        alert(message);
    }

    showLoading(show = true) {
        this.loading.style.display = show ? 'flex' : 'none';

        if (!show) {
            const hasComments = this.commentsContainer.querySelector('.comment');
            this.noComments.style.display = !hasComments ? 'block' : 'none';
        } else {
            this.noComments.style.display = 'none';
        }
    }

    async loadComments() {
        this.showLoading(true);
        const offset = (this.currentPage - 1) * this.limit;

        let url;
        if (this.isSearchMode) {
            url = `${this.apiUrl}/search?query=${encodeURIComponent(this.currentQuery)}&limit=${this.limit}&offset=${offset}`;
        } else {
            url = `${this.apiUrl}?limit=${this.limit}&offset=${offset}&sort=${this.currentSort}`;
        }

        try {
            const comments = await this.apiCall(url);
            this.renderComments(comments || []);
            this.hasMore = comments && comments.length === this.limit;
            this.updatePagination();
        } catch (error) {
            this.renderComments([]);
            this.hasMore = false;
            this.updatePagination();
        } finally {
            this.showLoading(false);
        }
    }

    renderComments(comments) {
        // Clear existing comments but keep loading and no-comments elements
        const existingComments = this.commentsContainer.querySelectorAll('.comment');
        existingComments.forEach(comment => comment.remove());

        if (!comments.length) {
            return;
        }

        comments.forEach(comment => {
            this.renderComment(comment, 0, this.commentsContainer);
        });
    }

    renderComment(comment, level, container) {
        const commentEl = document.createElement('div');
        commentEl.className = 'comment';
        commentEl.dataset.id = comment.id;

        const totalChildren = this.countChildren(comment);
        const isCollapsed = this.collapsedComments.has(comment.id);
        const content = comment.deleted ? '[Комментарий удален]' : comment.content;
        const isDeleted = comment.deleted;

        commentEl.innerHTML = `
            <div class="comment-wrapper ${isDeleted ? 'deleted-comment' : ''}">
                <div class="comment-header">
                    <div class="comment-meta">
                        ${totalChildren > 0 ?
            `<button class="collapse-btn ${isCollapsed ? 'collapsed' : ''}" data-id="${comment.id}">
                                ${isCollapsed ? '▶' : '▼'}
                            </button>` :
            '<span class="collapse-spacer">•</span>'
        }
                        <span class="comment-author">${this.escapeHtml(comment.author)}</span>
                        <span class="comment-date">${this.formatDate(comment.created_at)}</span>
                        ${totalChildren > 0 ?
            `<span class="children-count">${totalChildren} ${this.getChildrenText(totalChildren)}</span>` :
            ''
        }
                    </div>
                    <div class="comment-actions">
                        ${!isDeleted ? `
                            <button class="reply-btn" data-id="${comment.id}" data-author="${this.escapeHtml(comment.author)}">
                                Ответить
                            </button>
                            <button class="delete-btn" data-id="${comment.id}">
                                Удалить
                            </button>
                        ` : ''}
                    </div>
                </div>
                <div class="comment-content">${this.escapeHtml(content)}</div>
            </div>
        `;

        // Attach event listeners for this comment
        if (!isDeleted) {
            const replyBtn = commentEl.querySelector('.reply-btn');
            const deleteBtn = commentEl.querySelector('.delete-btn');

            if (replyBtn) {
                replyBtn.addEventListener('click', () => {
                    this.openReplyModal(comment.id, comment.author, comment.content);
                });
            }

            if (deleteBtn) {
                deleteBtn.addEventListener('click', () => {
                    this.deleteComment(comment.id);
                });
            }
        }

        const collapseBtn = commentEl.querySelector('.collapse-btn');
        if (collapseBtn) {
            collapseBtn.addEventListener('click', () => {
                this.toggleCollapse(comment.id);
            });
        }

        // Create children container
        const childrenContainer = document.createElement('div');
        childrenContainer.className = 'children-container';
        commentEl.appendChild(childrenContainer);

        container.appendChild(commentEl);

        // Recursively render children if not collapsed
        if (comment.children && comment.children.length > 0 && !isCollapsed) {
            comment.children.forEach(child => {
                this.renderComment(child, level + 1, childrenContainer);
            });
        }
    }

    countChildren(comment) {
        if (!comment.children || !comment.children.length) return 0;

        let count = comment.children.length;
        comment.children.forEach(child => {
            count += this.countChildren(child);
        });

        return count;
    }

    getChildrenText(count) {
        if (count % 10 === 1 && count % 100 !== 11) return 'ответ';
        if ([2, 3, 4].includes(count % 10) && ![12, 13, 14].includes(count % 100)) return 'ответа';
        return 'ответов';
    }

    toggleCollapse(id) {
        if (this.collapsedComments.has(id)) {
            this.collapsedComments.delete(id);
        } else {
            this.collapsedComments.add(id);
        }
        this.loadComments();
    }

    async createComment(parentId = null) {
        const author = parentId ? this.replyAuthorInput.value.trim() : this.authorInput.value.trim();
        const content = parentId ? this.replyContentInput.value.trim() : this.contentInput.value.trim();

        if (!author || !content) {
            this.showError('Заполните все поля');
            return;
        }

        const payload = { author, content };
        if (parentId) {
            payload.parent_id = parentId;
        }

        try {
            await this.apiCall(this.apiUrl, {
                method: 'POST',
                body: JSON.stringify(payload)
            });

            if (parentId) {
                this.closeReplyModal();
            } else {
                this.authorInput.value = '';
                this.contentInput.value = '';
            }

            this.loadComments();
        } catch (error) {
            // Error already handled in apiCall
        }
    }

    async deleteComment(id) {
        if (!confirm('Удалить комментарий?')) return;

        try {
            await this.apiCall(`${this.apiUrl}/${id}`, { method: 'DELETE' });
            this.loadComments();
        } catch (error) {
            // Error already handled in apiCall
        }
    }

    async search() {
        const query = this.searchInput.value.trim();
        if (!query) {
            this.showError('Введите запрос для поиска');
            return;
        }

        this.isSearchMode = true;
        this.currentQuery = query;
        this.currentPage = 1;
        this.collapsedComments.clear();
        this.loadComments();
    }

    clearSearch() {
        this.searchInput.value = '';
        this.isSearchMode = false;
        this.currentQuery = '';
        this.currentPage = 1;
        this.collapsedComments.clear();
        this.loadComments();
    }

    openReplyModal(id, author, content) {
        this.replyToId = id;
        const truncatedContent = content.length > 100 ? content.slice(0, 100) + '...' : content;
        this.replyTo.textContent = `Ответ на ${author}: "${truncatedContent}"`;
        this.replyAuthorInput.value = '';
        this.replyContentInput.value = '';
        this.replyModal.style.display = 'block';
        this.replyAuthorInput.focus();
    }

    closeReplyModal() {
        this.replyModal.style.display = 'none';
        this.replyToId = null;
    }

    async submitReply() {
        if (this.replyToId) {
            await this.createComment(this.replyToId);
        }
    }

    updatePagination() {
        this.pageInfo.textContent = `Страница ${this.currentPage}`;
        this.prevBtn.disabled = this.currentPage <= 1;
        this.nextBtn.disabled = !this.hasMore;
    }

    prevPage() {
        if (this.currentPage > 1) {
            this.currentPage--;
            this.loadComments();
        }
    }

    nextPage() {
        if (this.hasMore) {
            this.currentPage++;
            this.loadComments();
        }
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString('ru-RU', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new CommentTree();
});