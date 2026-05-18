use eframe::egui;

mod config;
mod model;
mod ui;
use crate::ui::NotiflowConfigUI;

fn main() {
    let native_options = eframe::NativeOptions {
        persist_window: false,
        viewport: egui::ViewportBuilder::default()
            .with_resizable(false)
            .with_maximized(false)
            .with_visible(true)
            .with_fullscreen(false)
            .with_maximize_button(false)
            .with_inner_size([550.0, 700.0]),
        ..Default::default()
    };
    let data = config::load().unwrap_or_default();
    let _ = eframe::run_native(
        "Configuracion Notiflow",
        native_options,
        Box::new(|cc| Ok(Box::new(NotiflowConfigUI::new(cc, data)))),
    );
}
