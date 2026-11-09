import { dayOf, type Day } from './days';

// Today's date, kept current: it moves on at midnight and when the phone
// wakes the app up. Anything that shows "today" reads it from here.
class Today {
  day = $state<Day>(dayOf(new Date()));

  #users = 0;
  #stop: (() => void) | null = null;

  /** Starts following the date; call the function it returns to stop. */
  follow(): () => void {
    if (this.#users++ === 0) {
      const tick = () => (this.day = dayOf(new Date()));
      tick();
      const timer = setInterval(tick, 60_000);
      document.addEventListener('visibilitychange', tick);
      this.#stop = () => {
        clearInterval(timer);
        document.removeEventListener('visibilitychange', tick);
      };
    }
    let stopped = false;
    return () => {
      if (stopped) return;
      stopped = true;
      if (--this.#users === 0) {
        this.#stop?.();
        this.#stop = null;
      }
    };
  }
}

export const today = new Today();
