import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";

dayjs.extend(relativeTime);

export const formatTimeAgo = (dateString: string): string => {
  return dayjs(dateString).fromNow(); // e.g. "5 hours ago", "2 days ago"
};
