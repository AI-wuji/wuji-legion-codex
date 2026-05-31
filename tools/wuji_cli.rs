use std::env;
use std::fs;
use std::path::{Path, PathBuf};

fn usage() {
    eprintln!("Usage:");
    eprintln!("  wuji-cli pptx-preflight --workspace <dir> [--generator <file>]");
    eprintln!("  wuji-cli pptx-batch-gate --workspace <dir> [--generator <file>]");
}

fn arg_value(args: &[String], name: &str) -> Option<String> {
    args.windows(2)
        .find(|pair| pair[0] == name)
        .map(|pair| pair[1].clone())
}

fn existing_plan_file(workspace: &Path, stem: &str) -> Option<PathBuf> {
    ["json", "md", "txt"]
        .iter()
        .map(|ext| workspace.join(format!("{stem}.{ext}")))
        .find(|path| path.exists())
}

fn non_empty(path: &Path) -> bool {
    fs::metadata(path).map(|m| m.len() >= 20).unwrap_or(false)
}

fn scan_generator(path: &Path) -> Vec<String> {
    let mut failures = Vec::new();
    let Ok(text) = fs::read_to_string(path) else {
        failures.push(format!("generator_unreadable={}", path.display()));
        return failures;
    };
    let lower = text.to_lowercase();

    let full_slide_picture =
        lower.contains("add_picture")
            && (lower.contains("slide_width") || lower.contains("slide_height"))
            && (lower.contains("0, 0") || lower.contains("left=0") || lower.contains("top=0"));
    if full_slide_picture {
        failures.push("generator_looks_like_full_slide_picture_route".to_string());
    }

    let raster_slide_route =
        lower.contains("full-slide-image")
            || lower.contains("full_slide_image")
            || lower.contains("slide as image")
            || lower.contains("render entire slide")
            || lower.contains("每页一张")
            || (lower.contains("image.new") && lower.contains("canvas.save"));
    if raster_slide_route {
        failures.push("generator_looks_like_raster_slide_route".to_string());
    }

    failures
}

fn pptx_preflight(args: &[String]) -> i32 {
    let Some(workspace) = arg_value(args, "--workspace") else {
        usage();
        return 2;
    };
    let workspace = PathBuf::from(workspace);
    let mut failures = Vec::new();

    if !workspace.is_dir() {
        failures.push(format!("workspace_missing={}", workspace.display()));
    }

    for stem in ["reference-frame-map", "reusable-asset-map", "illustration-plan"] {
        match existing_plan_file(&workspace, stem) {
            Some(path) if non_empty(&path) => {}
            Some(path) => failures.push(format!("plan_file_too_small={}", path.display())),
            None => failures.push(format!("missing_required_plan={stem}")),
        }
    }

    if let Some(generator) = arg_value(args, "--generator") {
        failures.extend(scan_generator(Path::new(&generator)));
    }

    if failures.is_empty() {
        println!("GO pptx-preflight");
        0
    } else {
        println!("NO-GO pptx-preflight");
        for failure in failures {
            println!("- {failure}");
        }
        1
    }
}

fn existing_pilot_file(workspace: &Path, stem: &str) -> Option<PathBuf> {
    ["png", "jpg", "jpeg", "pptx", "json", "md", "txt"]
        .iter()
        .map(|ext| workspace.join(format!("{stem}.{ext}")))
        .find(|path| path.exists())
}

fn pptx_batch_gate(args: &[String]) -> i32 {
    let mut failures = Vec::new();
    let preflight_code = {
        let Some(workspace) = arg_value(args, "--workspace") else {
            usage();
            return 2;
        };
        let workspace = PathBuf::from(workspace);

        if !workspace.is_dir() {
            failures.push(format!("workspace_missing={}", workspace.display()));
        }

        for stem in ["reference-frame-map", "reusable-asset-map", "illustration-plan"] {
            match existing_plan_file(&workspace, stem) {
                Some(path) if non_empty(&path) => {}
                Some(path) => failures.push(format!("plan_file_too_small={}", path.display())),
                None => failures.push(format!("missing_required_plan={stem}")),
            }
        }

        for stem in ["pilot-preview", "pilot-page"] {
            match existing_pilot_file(&workspace, stem) {
                Some(path) if non_empty(&path) => {}
                Some(path) => failures.push(format!("pilot_file_too_small={}", path.display())),
                None => failures.push(format!("missing_required_pilot={stem}")),
            }
        }

        match existing_plan_file(&workspace, "pilot-score") {
            Some(path) if non_empty(&path) => {}
            Some(path) => failures.push(format!("pilot_score_too_small={}", path.display())),
            None => failures.push("missing_required_pilot=pilot-score".to_string()),
        }

        if let Some(generator) = arg_value(args, "--generator") {
            failures.extend(scan_generator(Path::new(&generator)));
        }

        if failures.is_empty() { 0 } else { 1 }
    };

    if preflight_code == 0 {
        println!("GO pptx-batch-gate");
        0
    } else {
        println!("NO-GO pptx-batch-gate");
        for failure in failures {
            println!("- {failure}");
        }
        1
    }
}

fn main() {
    let args: Vec<String> = env::args().collect();
    let Some(command) = args.get(1).map(String::as_str) else {
        usage();
        std::process::exit(2);
    };

    let code = match command {
        "pptx-preflight" => pptx_preflight(&args[2..]),
        "pptx-batch-gate" => pptx_batch_gate(&args[2..]),
        _ => {
            usage();
            2
        }
    };
    std::process::exit(code);
}
