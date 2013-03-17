all:
	git pack-objects --all-progress-implied --revs --stdout --delta-base-offset --progress < refs 1>/dev/zero
thin:
	git pack-objects --all-progress-implied --revs --thin --stdout --delta-base-offset --progress < refs 1>/dev/zero
