const $sidebar = $('#sidebar');
const $overlay = $('#overlay');
const $modal = $('#commonModal');
const FADE = 200;
let sidebarOpen = false;

function showOverlay() { $overlay.stop(true, true).fadeIn(FADE); }
function hideOverlay() { $overlay.stop(true, true).fadeOut(FADE); }

function openSidebar() {
    closeModal();
    $sidebar.stop(true).animate({ left: 0 }, 300, function () {
        sidebarOpen = true;
    });
    showOverlay();
}

function closeSidebar(callback) {
    $sidebar.stop(true, true).animate({ left: -250 }, 300, function () {
        if (typeof callback === 'function') callback();
        else hideOverlay();
        sidebarOpen = false;
    });
}
function toggleMenu() { sidebarOpen ? closeSidebar() : openSidebar(); }

function openModal(html, buttons) {
    closeSidebar(function () {
        $modal.find('.modal-body').html(html || '');
        const $footer = $modal.find('.modal-footer').empty();

        if (buttons && buttons.length) {
            $footer.show();
            buttons.forEach(btn => {
                $('<button>')
                    .text(btn.text)
                    .addClass(btn.class || '')
                    .on('click', () => { if (btn.click) btn.click(); closeModal(); })
                    .appendTo($footer);
            });
        } else { $footer.hide(); }

        showOverlay();
        $modal.stop(true, true).fadeIn(FADE);
    });
}
function closeModal() {
    $modal.find('.modal-body').empty();
    $modal.find('.modal-footer').empty().hide();
    $modal.stop(true, true).fadeOut(FADE, hideOverlay);
}

// API
function Modal(opt) { openModal(opt.html, opt.buttons || []); }
function showAlert(msg) { Modal({ html: `<p>${msg}</p>`, buttons: [{ text: '닫기' }] }); }
function showConfirm(msg, onOk, onCancel) {
    Modal({
        html: `<p>${msg}</p>`, buttons: [
            { text: '확인', click: onOk },
            { text: '취소', click: onCancel }
        ]
    });
}
function loadPageModal() { $.get('page.html', data => Modal({ html: data })); }
function showToast(msg) {
    Modal({ html: `<p>${msg}</p>` });
    setTimeout(closeModal, 2000);
}
