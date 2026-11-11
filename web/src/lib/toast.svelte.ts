// A short note at the bottom of the screen, like "Moved to Tomorrow · Undo".
// One shows at a time; a new one takes the place of the last.

export interface ToastAction {
  label: string;
  run: () => unknown;
}

export interface Toast {
  id: number;
  text: string;
  action?: ToastAction;
}

class Toasts {
  current = $state<Toast | null>(null);
  #timer: ReturnType<typeof setTimeout> | undefined;
  #id = 0;

  show(text: string, action?: ToastAction, ms = 6000) {
    clearTimeout(this.#timer);
    const toast = { id: ++this.#id, text, action };
    this.current = toast;
    this.#timer = setTimeout(() => this.dismiss(toast.id), ms);
  }

  /** Hides the toast, or only the one with this id if it's still showing. */
  dismiss(id?: number) {
    if (id === undefined || this.current?.id === id) this.current = null;
  }
}

export const toasts = new Toasts();
