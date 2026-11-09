import { addDays, type Day } from '../days';
import { progress, type Check, type Habit } from './habits';

// A companion's mood follows its person's habits over the last few days:
// yesterday counts most, and what's done today only ever cheers it up, so
// the morning never starts sad because nothing is ticked yet. It gets
// droopy after days of skipping, never worse, and one good day brings it
// back.

export type Mood = 'glowing' | 'happy' | 'okay' | 'droopy' | 'sad';

export const MOODS: Record<Mood, { label: string; lines: string[] }> = {
  glowing: {
    label: 'Glowing',
    lines: ['Best day ever!', 'I could do a cartwheel.', 'Look at us go!'],
  },
  happy: {
    label: 'Happy',
    lines: ['This is nice.', 'I like our days.', "Let's keep it up!"],
  },
  okay: {
    label: 'Okay',
    lines: ['Hi! What shall we do today?', 'Ready when you are.', 'A little something today?'],
  },
  droopy: {
    label: 'A bit droopy',
    lines: ['I missed you yesterday.', 'Just one small thing?', 'Could use a little care.'],
  },
  sad: {
    label: 'Sad',
    lines: [
      "It's been a few quiet days.",
      'One small habit would cheer me up.',
      'I saved you a spot.',
    ],
  },
};

/** How a day went, 0 to 1, from its habits; null when nothing was due. */
export type DayScore = number | null;

/** A score from 0 (sad) to 1 (glowing). */
export function moodScore(habits: Habit[], checks: Check[], today: Day): number {
  const past = weighted([
    [progress(habits, checks, addDays(today, -1)), 0.6],
    [progress(habits, checks, addDays(today, -2)), 0.4],
  ]);
  const now = progress(habits, checks, today) ?? 0;
  // someone new (or back after a break with nothing due) starts happy
  const base = past ?? 0.7;
  return base + (1 - base) * 0.8 * now;
}

export function moodOf(score: number): Mood {
  if (score >= 0.9) return 'glowing';
  if (score >= 0.7) return 'happy';
  if (score >= 0.45) return 'okay';
  if (score >= 0.2) return 'droopy';
  return 'sad';
}

/** What it says today: the same line all day, so it doesn't flicker. */
export function lineFor(mood: Mood, today: Day, allDone: boolean): string {
  if (allDone)
    return mood === 'glowing' ? 'All done! You made my day.' : 'All done today, thank you!';
  const lines = MOODS[mood].lines;
  let h = 0;
  for (const c of today) h = (h * 31 + c.charCodeAt(0)) >>> 0;
  return lines[h % lines.length];
}

function weighted(days: [DayScore, number][]): number | null {
  const counted = days.filter((d): d is [number, number] => d[0] !== null);
  const total = counted.reduce((s, [, w]) => s + w, 0);
  return total ? counted.reduce((s, [v, w]) => s + v * w, 0) / total : null;
}
