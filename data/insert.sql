INSERT INTO years (Y, objective) VALUES
(2025, 'Launch new products and increase market presence');

INSERT INTO months (M, Y, objective) VALUES
(3, 2025, 'Launch new product features and gather feedback from early adopters'),
(4, 2025, 'Focus on customer feedback and optimize the user experience'),
(5, 2025, 'Plan and execute a mid-year review of product and marketing performance'),
(6, 2025, 'Develop new partnerships and expand the product’s reach in new markets');

INSERT INTO days (M, Y, dt, diary) VALUES
(3, 2025, '2025-03-01', 'Kickoff meeting for product feature launch and team planning'),
(3, 2025, '2025-03-25', 'Product feature launch and beta testing feedback collection'),
(4, 2025, '2025-04-10', 'Customer feedback analysis and prioritization of changes'),
(5, 2025, '2025-05-15', 'Mid-year review and strategy realignment session'),
(6, 2025, '2025-06-20', 'Negotiation with new partners for market expansion');


INSERT INTO task (title, dt, start_time, end_time, des, M, Y) VALUES
('Product feature launch planning', '2025-03-10', '09:00:00', '12:00:00', 'Plan and allocate resources for upcoming product feature launch', 3, 2025),
('Beta testing for new product features', '2025-03-25', '13:00:00', '16:00:00', 'Conduct beta testing and gather early feedback from users', 3, 2025),
('Customer feedback analysis', '2025-04-05', '10:00:00', '12:00:00', 'Analyze customer feedback to prioritize product improvements', 4, 2025),
('Marketing strategy optimization', '2025-05-01', '09:00:00', '12:00:00', 'Reevaluate marketing strategies and optimize campaigns', 5, 2025),
('Partnership development and negotiation', '2025-06-10', '14:00:00', '16:00:00', 'Engage in partnership discussions to expand market reach', 6, 2025);
