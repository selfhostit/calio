import type { RecordModel } from 'pocketbase';

// The people and places the household relies on.

export type Role = 'vet' | 'petsitter' | 'babysitter' | 'doctor' | 'neighbour' | 'other';

export interface Contact extends RecordModel {
  name: string;
  role: Role;
  phone: string;
  email: string;
  address: string;
  notes: string;
  created_by: string;
}

export const ROLES: { role: Role; label: string; plural: string; icon: string }[] = [
  { role: 'vet', label: 'Vet', plural: 'Vets', icon: 'paw' },
  { role: 'petsitter', label: 'Pet sitter', plural: 'Pet sitters', icon: 'paw' },
  { role: 'babysitter', label: 'Babysitter', plural: 'Babysitters', icon: 'user' },
  { role: 'doctor', label: 'Doctor', plural: 'Doctors', icon: 'user' },
  { role: 'neighbour', label: 'Neighbour', plural: 'Neighbours', icon: 'user' },
  { role: 'other', label: 'Other', plural: 'Others', icon: 'user' },
];

/** A phone number as a tel: link, keeping only what dialers understand. */
export function telLink(phone: string): string {
  return 'tel:' + phone.replace(/[^\d+*#,;]/g, '');
}

/** The contact whose name matches a sitter's name, if there is one. */
export function findByName(contacts: Contact[], name: string): Contact | undefined {
  const n = name.trim().toLowerCase();
  return n ? contacts.find((c) => c.name.trim().toLowerCase() === n) : undefined;
}
