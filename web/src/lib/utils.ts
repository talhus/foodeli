import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Formats minor units (cents / kuruş) into standard human-readable currency strings.
 * e.g., 1450 cents -> "€14.50"
 */
export function formatCurrency(cents: number, currency: string = "EUR"): string {
  const amount = (cents || 0) / 100;
  return new Intl.NumberFormat("en-IE", {
    style: "currency",
    currency: currency || "EUR",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
}

/**
 * Formats ISO date strings to localized readable timestamps.
 */
export function formatDate(dateString: string): string {
  if (!dateString) return "";
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return dateString;
    return new Intl.DateTimeFormat("en-GB", {
      dateStyle: "medium",
      timeStyle: "short",
    }).format(date);
  } catch {
    return dateString || "";
  }
}

