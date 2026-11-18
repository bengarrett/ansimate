(() => {
  const go = new Go();
  let wasmReady = false;
  let currentFileBytes = null;
  let currentFileName = "ansimation";
  let activeBlobUrl = null;

  const ansimateForm = document.getElementById("ansimateForm");
  const wasmStatus = document.getElementById("wasmStatus");
  const filePicker = document.getElementById("filePicker");
  const dropzone = document.getElementById("dropzone");
  const selectedFileInfo = document.getElementById("selectedFileInfo");
  const statusMessage = document.getElementById("statusMessage");

  const previewContainer = document.getElementById("previewContainer");
  const previewLoading = document.getElementById("previewLoading");
  const previewPlaceholder = previewContainer.querySelector(".preview-placeholder");
  const previewImage = document.getElementById("previewImage");
  const resultMeta = document.getElementById("resultMeta");
  const metaInputSize = document.getElementById("metaInputSize");
  const metaWebpSize = document.getElementById("metaWebpSize");
  const metaTime = document.getElementById("metaTime");
  const downloadBtn = document.getElementById("downloadBtn");

  // Format bytes into human readable string
  function formatBytes(bytes) {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(2) + " MB";
  }

  // Initialize WebAssembly
  async function initWasm() {
    try {
      const response = await fetch("ansimate-web.wasm");
      if (!response.ok) {
        throw new Error(`Failed to fetch ansimate-web.wasm (${response.status} ${response.statusText})`);
      }
      const { instance } = await WebAssembly.instantiateStreaming(response, go.importObject);
      go.run(instance);
      wasmReady = true;
      wasmStatus.textContent = "Ready v" + globalThis.ansimate.version();
      if (currentFileBytes) triggerConvert();
    } catch (err) {
      console.error("WASM load error:", err);
      wasmStatus.textContent = "WASM Failed";
      showStatus("Could not load WebAssembly module: " + err.message, "error");
    }
  }

  function showStatus(msg, type = "error") {
    statusMessage.textContent = msg;
    statusMessage.className = `status-message ${type}`;
    statusMessage.classList.remove("hidden");
  }

  function clearStatus() {
    statusMessage.classList.add("hidden");
    statusMessage.textContent = "";
  }

  // Handle file selection
  function handleFile(file) {
    clearStatus();
    currentFileName = file.name.replace(/\.[^/.]+$/, "");
    const reader = new FileReader();
    reader.onload = (e) => {
      currentFileBytes = new Uint8Array(e.target.result);
      selectedFileInfo.textContent = `Selected: ${file.name} (${formatBytes(currentFileBytes.length)})`;
      selectedFileInfo.classList.remove("hidden");
      if (wasmReady) {
        triggerConvert();
      }
    };
    reader.onerror = () => showStatus("Failed to read file", "error");
    reader.readAsArrayBuffer(file);
  }

  // Drag and drop events
  dropzone.addEventListener("dragover", (e) => {
    e.preventDefault();
    dropzone.classList.add("dragover");
  });
  dropzone.addEventListener("dragleave", () => dropzone.classList.remove("dragover"));
  dropzone.addEventListener("drop", (e) => {
    e.preventDefault();
    dropzone.classList.remove("dragover");
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFile(e.dataTransfer.files[0]);
    }
  });
  filePicker.addEventListener("change", (e) => {
    if (e.target.files && e.target.files.length > 0) {
      handleFile(e.target.files[0]);
    }
  });

  // Perform conversion
  async function triggerConvert() {
    if (!wasmReady || !currentFileBytes) return;

    clearStatus();
    previewLoading.classList.remove("hidden");

    // Allow the browser to render the loading overlay and spin before WASM blocks the thread
    await new Promise((resolve) => {
      setTimeout(resolve, 30);
    });

    const options = {
      palette: document.getElementById("palette").value,
      baudRate: parseInt(document.getElementById("baudRate").value, 10),
      fps: parseInt(document.getElementById("fps").value, 10),
      delay: parseFloat(document.getElementById("delay").value) * 1000,
      cga: document.getElementById("cga").checked,
    };

    const startTime = performance.now();

    try {
      const webpBytes = await globalThis.ansimate.convert(currentFileBytes, options);
      const elapsed = ((performance.now() - startTime) / 1000).toFixed(2);

      if (activeBlobUrl) {
        URL.revokeObjectURL(activeBlobUrl);
      }
      const blob = new Blob([webpBytes], { type: "image/webp" });
      activeBlobUrl = URL.createObjectURL(blob);

      previewImage.src = activeBlobUrl;
      previewImage.classList.remove("hidden");
      previewPlaceholder.classList.add("hidden");

      metaInputSize.textContent = formatBytes(currentFileBytes.length);
      metaWebpSize.textContent = formatBytes(webpBytes.length);
      metaTime.textContent = `${elapsed}s`;
      downloadBtn.href = activeBlobUrl;
      downloadBtn.download = `${currentFileName}.webp`;
      resultMeta.classList.remove("hidden");
    } catch (err) {
      showStatus("Conversion error: " + err.message, "error");
    } finally {
      previewLoading.classList.add("hidden");
    }
  }

  ansimateForm.addEventListener("submit", (e) => e.preventDefault());

  ansimateForm.addEventListener("change", () => {
    if (currentFileBytes && wasmReady) {
      triggerConvert();
    }
  });

  ansimateForm.addEventListener("reset", () => {
    clearStatus();
    currentFileBytes = null;
    currentFileName = "ansimation";

    selectedFileInfo.classList.add("hidden");
    selectedFileInfo.textContent = "";

    previewLoading.classList.add("hidden");

    if (activeBlobUrl) {
      URL.revokeObjectURL(activeBlobUrl);
      activeBlobUrl = null;
    }
    previewImage.src = "";
    previewImage.classList.add("hidden");
    previewPlaceholder.classList.remove("hidden");
    resultMeta.classList.add("hidden");
    downloadBtn.removeAttribute("href");
  });

  // Start initialization
  initWasm();
})();
