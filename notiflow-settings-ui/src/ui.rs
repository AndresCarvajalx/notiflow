use egui::{RichText, scroll_area::ScrollArea};
use egui_file_dialog::FileDialog;

use crate::{config, model::Data};

#[derive(Default)]
pub struct NotiflowConfigUI {
    data: Data,
    file_dialog: FileDialog,
    save_status: SaveStatus,
}

#[derive(Default, Clone, Copy, PartialEq)]
enum SaveStatus {
    #[default]
    None,
    Saving,
    Success,
    Error,
}

impl NotiflowConfigUI {
    pub fn new(_: &eframe::CreationContext<'_>, data: Data) -> Self {
        Self {
            data,
            file_dialog: FileDialog::new(),
            save_status: SaveStatus::None,
        }
    }

    fn create_labeled_row<F>(ui: &mut egui::Ui, label: &str, add_content: F)
    where
        F: FnOnce(&mut egui::Ui),
    {
        ui.horizontal(|ui| {
            ui.label(RichText::new(label).size(12.0));
            ui.with_layout(egui::Layout::right_to_left(egui::Align::Center), |ui| {
                add_content(ui);
            });
        });
    }

    fn section_title(ui: &mut egui::Ui, title: &str) {
        ui.vertical(|ui| {
            ui.add_space(8.0);
            ui.heading(RichText::new(title).size(18.0).color(egui::Color32::from_rgb(70, 130, 180)));
            ui.separator();
            ui.add_space(4.0);
        });
    }

    fn on_save(&mut self) {
        self.save_status = SaveStatus::Saving;
        match config::save(&self.data) {
            Ok(_) => {
                self.save_status = SaveStatus::Success;
            }
            Err(_) => {
                self.save_status = SaveStatus::Error;
            }
        }
    }
}

impl eframe::App for NotiflowConfigUI {
    fn ui(&mut self, ui: &mut egui::Ui, _frame: &mut eframe::Frame) {
        egui::Panel::top("top").show_inside(ui, |ui| {
            ui.add_space(8.0);
            ui.horizontal(|ui| {
                ui.add_space(12.0);
                let title = RichText::new("🔩 CONFIGURACIÓN NOTIFLOW").size(24.0).strong();
                ui.heading(title);
            });
            ui.add_space(8.0);
            ui.separator();
        });

        egui::Panel::bottom("bottom").show_inside(ui, |ui| {
            ui.separator();
            ui.add_space(6.0);
            ui.horizontal(|ui| {
                ui.add_space(12.0);
                
                let button = egui::Button::new(RichText::new("💾 Guardar").size(12.0))
                    .fill(egui::Color32::from_rgb(70, 130, 180))
                    .min_size(egui::Vec2::new(120.0, 32.0));
                    
                if ui.add(button).clicked() {
                    self.on_save();
                }

                match self.save_status {
                    SaveStatus::Success => {
                        ui.label(RichText::new("✅ Guardado exitosamente").size(11.0).color(egui::Color32::GREEN));
                    }
                    SaveStatus::Error => {
                        ui.label(RichText::new("❌ Error al guardar").size(11.0).color(egui::Color32::RED));
                    }
                    _ => {}
                }
            });
            ui.add_space(6.0);
        });

        egui::CentralPanel::default().show_inside(ui, |ui| {
            ScrollArea::vertical()
                .auto_shrink([false; 2])
                .show(ui, |ui| {
                    ui.vertical_centered(|ui| {
                        ui.add_space(8.0);
                        
                        ui.horizontal(|ui| {
                            ui.add_space(20.0);
                            
                            ui.vertical(|ui| {
                                ui.set_max_width(500.0);

                                NotiflowConfigUI::section_title(ui, "📊 Excel");
                                ui.add_space(2.0);

                                NotiflowConfigUI::create_labeled_row(ui, "Ruta del archivo:", |ui| {
                                    if ui.button("📁").on_hover_text("Seleccionar archivo").clicked() {
                                        self.file_dialog.pick_file();
                                    }
                                    ui.text_edit_singleline(&mut self.data.excel.path)
                                        .on_hover_text("Ruta del archivo Excel");
                                });

                                self.file_dialog.update(ui.ctx());

                                if let Some(path) = self.file_dialog.take_picked() {
                                    self.data.excel.path = path.display().to_string();
                                }

                                NotiflowConfigUI::create_labeled_row(ui, "Fila de encabezado:", |ui| {
                                    ui.add(egui::Slider::new(&mut self.data.excel.header_row, 0..=20)
                                        .show_value(true)
                                        .fixed_decimals(0)
                                        .max_decimals(0))
                                        .on_hover_text("Número de fila donde inician los encabezados");
                                });

                                ui.add_space(12.0);

                                NotiflowConfigUI::section_title(ui, "📋 Columnas Excel");

                                let campos_excel: [(&str, &mut String); 11] = [
                                    ("Tipo Transacción", &mut self.data.columnas.tipo_transaccion),
                                    ("Cliente", &mut self.data.columnas.cliente),
                                    ("Placa", &mut self.data.columnas.placa),
                                    ("Valor Actual", &mut self.data.columnas.valor_actual),
                                    ("Porcentaje Interés", &mut self.data.columnas.porcentaje_interes),
                                    ("Valor Interés Mensual", &mut self.data.columnas.valor_interes_mensual),
                                    ("Vencimiento Interés", &mut self.data.columnas.vencimiento_interes),
                                    ("Días Corridos", &mut self.data.columnas.dias_corridos),
                                    ("Valor Interés Vencido", &mut self.data.columnas.valor_interes_vencido),
                                    ("Saldo Actual", &mut self.data.columnas.saldo_actual),
                                    ("Teléfono", &mut self.data.columnas.telefono),
                                ];

                                for (label, campo) in campos_excel {
                                    NotiflowConfigUI::create_labeled_row(ui, label, |ui| {
                                        ui.text_edit_singleline(campo);
                                    });
                                    ui.add_space(1.0);
                                }

                                ui.add_space(12.0);

                                NotiflowConfigUI::section_title(ui, "💬 WhatsApp");

                                let campos_whatsapp = [
                                    ("Token de acceso", &mut self.data.whatsapp.token),
                                    ("ID del teléfono", &mut self.data.whatsapp.phone_id),
                                    ("Código de país", &mut self.data.whatsapp.codigo_pais),
                                ];

                                for (label, campo) in campos_whatsapp {
                                    NotiflowConfigUI::create_labeled_row(ui, label, |ui| {
                                        ui.text_edit_singleline(campo);
                                    });
                                    ui.add_space(1.0);
                                }

                                ui.add_space(12.0);

                                NotiflowConfigUI::section_title(ui, "⏰ Programador");

                                NotiflowConfigUI::create_labeled_row(ui, "Días de vencimiento:", |ui| {
                                    ui.add(egui::Slider::new(
                                        &mut self.data.scheduler.dias_vencimiento,
                                        0..=60,
                                    )
                                    .show_value(true)
                                    .fixed_decimals(0)
                                    .max_decimals(0))
                                        .on_hover_text("Días para considerar como vencido");
                                });

                                ui.add_space(12.0);

                                NotiflowConfigUI::section_title(ui, "🖥️ Servidor");

                                NotiflowConfigUI::create_labeled_row(ui, "Puerto:", |ui| {
                                    ui.add(egui::Slider::new(&mut self.data.server.port, 1000..=9999)
                                        .show_value(true)
                                        .fixed_decimals(0)
                                        .max_decimals(0))
                                        .on_hover_text("Puerto en el que correrá el servidor");
                                });

                                ui.add_space(20.0);
                            });
                            
                            ui.add_space(20.0);
                        });
                    });
                });
        });
    }
}
