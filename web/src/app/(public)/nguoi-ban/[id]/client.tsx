"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, MapPin, MessageCircle, Star, Flag, Calendar } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import {
  getPublicProfile,
  getSellerRatings,
  getRatingSummary,
  createRating,
  createConversation,
  createReport,
  type PublicProfile,
  type Rating,
  type RatingSummary,
} from "@/services/api";
import { useAuth } from "@/lib/auth";
import { formatDate, timeAgo } from "@/lib/utils";
import { toast } from "sonner";

export default function SellerProfilePage() {
  const { id } = useParams<{ id: string }>();
  const { user, token, hasPermission } = useAuth();
  const router = useRouter();
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [ratings, setRatings] = useState<Rating[]>([]);
  const [summary, setSummary] = useState<RatingSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [contacting, setContacting] = useState(false);

  // Rating form
  const [ratingStars, setRatingStars] = useState(5);
  const [ratingComment, setRatingComment] = useState("");
  const [submittingRating, setSubmittingRating] = useState(false);
  const [hasRated, setHasRated] = useState(false);

  // Report
  const [showReport, setShowReport] = useState(false);
  const [reportReason, setReportReason] = useState("");
  const [reportDesc, setReportDesc] = useState("");
  const [submittingReport, setSubmittingReport] = useState(false);

  const isOwnProfile = user?.id === id;

  useEffect(() => {
    if (!id) return;
    Promise.all([
      getPublicProfile(id, token || undefined),
      getSellerRatings(id, 1, 20, token || undefined),
      getRatingSummary(id, token || undefined),
    ])
      .then(([p, r, s]) => {
        setProfile(p);
        setRatings(r.data);
        setSummary(s);
        // Check if current user already rated
        if (user) {
          setHasRated(r.data.some((rating) => rating.reviewer_id === user.id));
        }
      })
      .catch(() => toast.error("Không tìm thấy người bán"))
      .finally(() => setLoading(false));
  }, [id, user, token]);

  async function handleContact() {
    if (!hasPermission("chat.send")) {
      toast.error(token ? "Không có quyền thực hiện - Cần gia hạn gói dịch vụ" : "Đăng nhập để tiếp tục");
      if (!token) router.push("/dang-nhap");
      return;
    }
    if (!id) return;
    setContacting(true);
    try {
      const conv = await createConversation(token!, id);
      router.push(`/tin-nhan/${conv.id}`);
    } catch (err) {
      if ((err as { status?: number })?.status === 403) {
        toast.error("Không có quyền thực hiện - Cần gia hạn gói dịch vụ");
      } else {
        toast.error(err instanceof Error ? err.message : "Không thể liên hệ");
      }
    } finally {
      setContacting(false);
    }
  }

  async function handleSubmitRating(e: React.FormEvent) {
    e.preventDefault();
    if (!hasPermission("ratings.create")) {
      toast.error(token ? "Không có quyền thực hiện - Cần gia hạn gói dịch vụ" : "Đăng nhập để tiếp tục");
      if (!token) router.push("/dang-nhap");
      return;
    }
    if (!id) return;
    if (!ratingComment.trim()) {
      toast.error("Vui lòng nhập nhận xét");
      return;
    }
    setSubmittingRating(true);
    try {
      const newRating = await createRating(token!, id, ratingStars, ratingComment.trim());
      setRatings((prev) => [newRating, ...prev]);
      setHasRated(true);
      setRatingComment("");
      // Refresh summary
      const s = await getRatingSummary(id);
      setSummary(s);
      toast.success("Đánh giá thành công");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Đánh giá thất bại");
    } finally {
      setSubmittingRating(false);
    }
  }

  async function handleReport(e: React.FormEvent) {
    e.preventDefault();
    if (!hasPermission("reports.create")) {
      toast.error(token ? "Không có quyền thực hiện - Cần gia hạn gói dịch vụ" : "Đăng nhập để tiếp tục");
      if (!token) router.push("/dang-nhap");
      return;
    }
    if (!id) return;
    if (!reportReason) {
      toast.error("Vui lòng chọn lý do báo cáo");
      return;
    }
    setSubmittingReport(true);
    try {
      await createReport(token!, "user", id, reportReason, reportDesc || undefined);
      toast.success("Đã gửi báo cáo");
      setShowReport(false);
      setReportReason("");
      setReportDesc("");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Báo cáo thất bại");
    } finally {
      setSubmittingReport(false);
    }
  }

  if (loading) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6 space-y-4">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-48 w-full rounded-lg" />
        <Skeleton className="h-60 w-full rounded-lg" />
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-12 text-center">
        <p className="text-muted-foreground">Không tìm thấy người bán</p>
        <Link href="/san-giao-dich">
          <Button variant="outline" className="mt-4">Quay lại sàn</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-6">
      <Link href="/san-giao-dich" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4">
        <ArrowLeft className="h-4 w-4" />
        Quay lại
      </Link>

      {/* Profile card */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex items-start gap-4">
            <Avatar className="h-16 w-16">
              {profile.avatar_url && <AvatarImage src={profile.avatar_url} alt={profile.name || "Avatar"} />}
              <AvatarFallback className="bg-primary/10 text-primary font-bold text-xl">
                {(profile.name || "?").charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1">
              <h1 className="text-xl font-bold">{profile.name || "Ẩn danh"}</h1>
              {profile.org_name && (
                <p className="text-sm text-muted-foreground">{profile.org_name}</p>
              )}
              {profile.province && (
                <p className="text-sm text-muted-foreground flex items-center gap-1 mt-1">
                  <MapPin className="h-3.5 w-3.5" />
                  {profile.province}
                </p>
              )}
              <p className="text-xs text-muted-foreground flex items-center gap-1 mt-1">
                <Calendar className="h-3 w-3" />
                Thành viên từ {formatDate(profile.created_at)}
              </p>
              {profile.description && (
                <p className="text-sm mt-3 whitespace-pre-wrap">{profile.description}</p>
              )}
            </div>
            <div className="flex flex-col gap-2">
              {!isOwnProfile && (
                <>
                  <Button className="gap-1.5" onClick={handleContact} disabled={contacting}>
                    <MessageCircle className="h-4 w-4" />
                    {contacting ? "Đang xử lý..." : "Nhắn tin"}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-1.5 text-destructive"
                    onClick={() => setShowReport(!showReport)}
                  >
                    <Flag className="h-3.5 w-3.5" />
                    Báo cáo
                  </Button>
                </>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Report form */}
      {showReport && (
        <Card className="mb-6">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Báo cáo người dùng</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleReport} className="space-y-3">
              <select
                value={reportReason}
                onChange={(e) => setReportReason(e.target.value)}
                className="w-full h-10 rounded-md border border-input bg-background px-3 text-sm"
                required
              >
                <option value="">Chọn lý do</option>
                <option value="fraud">Lừa đảo</option>
                <option value="false_info">Thông tin sai lệch</option>
                <option value="spam">Spam</option>
                <option value="other">Khác</option>
              </select>
              <textarea
                value={reportDesc}
                onChange={(e) => setReportDesc(e.target.value)}
                placeholder="Mô tả chi tiết (không bắt buộc)"
                className="w-full min-h-16 rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
              <div className="flex gap-2">
                <Button type="submit" variant="destructive" size="sm" disabled={submittingReport}>
                  {submittingReport ? "Đang gửi..." : "Gửi báo cáo"}
                </Button>
                <Button type="button" variant="ghost" size="sm" onClick={() => setShowReport(false)}>
                  Hủy
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Rating summary */}
      <Card className="mb-6">
        <CardHeader className="pb-3">
          <CardTitle className="text-lg flex items-center gap-2">
            <Star className="h-5 w-5" />
            Đánh giá
            {summary && summary.total > 0 && (
              <span className="text-sm font-normal text-muted-foreground">
                ({summary.average.toFixed(1)}/5 - {summary.total} đánh giá)
              </span>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {summary && summary.total > 0 ? (
            <div className="flex items-center gap-4 mb-4">
              <div className="text-3xl font-bold text-primary">{summary.average.toFixed(1)}</div>
              <div className="flex-1 space-y-1">
                {[5, 4, 3, 2, 1].map((star) => {
                  const count = summary.distribution[String(star)] || 0;
                  const pct = summary.total > 0 ? (count / summary.total) * 100 : 0;
                  return (
                    <div key={star} className="flex items-center gap-2 text-xs">
                      <span className="w-3">{star}</span>
                      <Star className="h-3 w-3 text-yellow-500 fill-yellow-500" />
                      <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                        <div className="h-full bg-yellow-500 rounded-full" style={{ width: `${pct}%` }} />
                      </div>
                      <span className="w-6 text-right text-muted-foreground">{count}</span>
                    </div>
                  );
                })}
              </div>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground mb-4">Chưa có đánh giá</p>
          )}

          {/* Rating form */}
          {!isOwnProfile && !hasRated && hasPermission("ratings.create") && (
            <form onSubmit={handleSubmitRating} className="border-t pt-4 space-y-3">
              <p className="text-sm font-medium">Đánh giá người bán</p>
              <div className="flex gap-1">
                {[1, 2, 3, 4, 5].map((s) => (
                  <button
                    key={s}
                    type="button"
                    onClick={() => setRatingStars(s)}
                    className="p-1"
                  >
                    <Star
                      className={`h-6 w-6 ${
                        s <= ratingStars ? "text-yellow-500 fill-yellow-500" : "text-muted-foreground"
                      }`}
                    />
                  </button>
                ))}
              </div>
              <textarea
                value={ratingComment}
                onChange={(e) => setRatingComment(e.target.value)}
                placeholder="Nhận xét về người bán..."
                className="w-full min-h-16 rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
              <Button type="submit" size="sm" disabled={submittingRating}>
                {submittingRating ? "Đang gửi..." : "Gửi đánh giá"}
              </Button>
            </form>
          )}
          {hasRated && (
            <p className="text-sm text-muted-foreground border-t pt-3">Bạn đã đánh giá người bán này</p>
          )}

          {/* Reviews list */}
          {ratings.length > 0 && (
            <div className="border-t pt-4 mt-4 space-y-4">
              {ratings.map((r) => (
                <div key={r.id} className="flex gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback className="text-xs bg-primary/10 text-primary">
                      {(r.reviewer_name || "?").charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">{r.reviewer_name || "Ẩn danh"}</span>
                      <div className="flex">
                        {[1, 2, 3, 4, 5].map((s) => (
                          <Star
                            key={s}
                            className={`h-3 w-3 ${
                              s <= r.stars ? "text-yellow-500 fill-yellow-500" : "text-muted-foreground"
                            }`}
                          />
                        ))}
                      </div>
                      <span className="text-xs text-muted-foreground">{timeAgo(r.created_at)}</span>
                    </div>
                    <p className="text-sm mt-1">{r.comment}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
